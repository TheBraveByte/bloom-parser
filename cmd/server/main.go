package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	pb "github.com/TheBraveByte/bloom-parser/gen/go/document_parser/v1"
	"github.com/TheBraveByte/bloom-parser/internal/config"
	"github.com/TheBraveByte/bloom-parser/internal/ingest"
	"github.com/TheBraveByte/bloom-parser/internal/observability"
	"github.com/TheBraveByte/bloom-parser/internal/ocr"
	"github.com/TheBraveByte/bloom-parser/internal/parser"
	"github.com/TheBraveByte/bloom-parser/internal/powerbi"
	"github.com/TheBraveByte/bloom-parser/internal/service"
	"github.com/TheBraveByte/bloom-parser/internal/textract"
)

//go:embed web/index.html
var indexHTML []byte

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	slog.SetDefault(observability.NewLogger(cfg.LogFormat, cfg.LogLevel))

	shutdownTrace, err := observability.SetupTracing(cfg.OTelEnabled)
	if err != nil {
		return fmt.Errorf("tracing setup: %w", err)
	}
	defer shutdownTrace(context.Background())

	var pub *powerbi.Publisher
	if cfg.PowerBI.Configured() {
		pub = powerbi.NewPublisher(powerbi.NewClient(cfg.PowerBI, nil), cfg.PowerBI.WorkspaceID)
		slog.Info("power bi integration enabled")
	} else {
		slog.Info("power bi credentials not set; PublishToPowerBI will return FAILED_PRECONDITION")
	}

	engine := ocr.Configured(ocr.Options{Languages: cfg.OCRLanguages})
	slog.Info("ocr engine", "name", engine.Name(), "available", engine.Available())

	ing := ingest.New(engine, ingest.Limits{MaxBytes: cfg.MaxBytes, MaxPages: cfg.MaxPages})
	srv := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	handler := service.NewServer(parser.NewService(), ing, pub).WithDefaultOCR(cfg.OCRDefault)
	pb.RegisterDocumentParserServiceServer(srv, handler)
	reflection.Register(srv)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	gateway, err := newGateway(context.Background(), cfg.Port)
	if err != nil {
		return fmt.Errorf("gateway: %w", err)
	}
	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: rootHandler(gateway, textract.NewRunner(), int64(cfg.MaxBytes)),
	}

	errCh := make(chan error, 2)
	go func() { errCh <- srv.Serve(lis) }()
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http serve: %w", err)
		}
	}()
	slog.Info("grpc server listening", "addr", lis.Addr().String())
	slog.Info("http/rest gateway listening", "addr", httpSrv.Addr)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case s := <-sig:
		slog.Info("shutting down", "signal", s.String())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(ctx)
		srv.GracefulStop()
		return nil
	case err := <-errCh:
		return fmt.Errorf("serve: %w", err)
	}
}

func newGateway(ctx context.Context, grpcPort int) (http.Handler, error) {
	mux := runtime.NewServeMux()
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", grpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	if err := pb.RegisterDocumentParserServiceHandler(ctx, mux, conn); err != nil {
		return nil, err
	}
	return mux, nil
}

func rootHandler(gateway http.Handler, tr *textract.Runner, maxBytes int64) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/v1/", gateway)
	mux.HandleFunc("POST /v1/table-extract", tableExtractHandler(tr, maxBytes))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexHTML)
	})
	return mux
}

const maxTableFiles = 64

func tableExtractHandler(tr *textract.Runner, maxBytes int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		if err := r.ParseMultipartForm(maxBytes); err != nil {
			http.Error(w, "multipart form: "+err.Error(), http.StatusBadRequest)
			return
		}
		heads := r.MultipartForm.File["file"]
		if len(heads) == 0 {
			http.Error(w, `missing "file" field`, http.StatusBadRequest)
			return
		}
		if len(heads) > maxTableFiles {
			http.Error(w, fmt.Sprintf("too many files (max %d)", maxTableFiles),
				http.StatusBadRequest)
			return
		}
		files := make([]textract.File, 0, len(heads))
		for _, h := range heads {
			f, err := h.Open()
			if err != nil {
				http.Error(w, "open upload: "+err.Error(), http.StatusBadRequest)
				return
			}
			data, err := io.ReadAll(f)
			f.Close()
			if err != nil {
				http.Error(w, "read upload: "+err.Error(), http.StatusBadRequest)
				return
			}
			files = append(files, textract.File{Name: h.Filename, Data: data})
		}
		res := tr.RunAll(r.Context(), files,
			textract.Options{Refine: r.FormValue("refine") == "true"})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}
}
