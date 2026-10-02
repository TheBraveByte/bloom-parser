from tableextract.normalize import (
    detect_header,
    facts_for_sheet,
    normalize_value,
    parse_period,
)


def test_parse_period_formats():
    assert parse_period("31-03-2018") == "2018-03-31"
    assert parse_period("31.03.2017 Audited") == "2017-03-31"
    assert parse_period("March 31, 2018") == "2018-03-31"
    assert parse_period("As at March 2018") == "2018-03"
    assert parse_period("FY 2018") == "2018"
    assert parse_period("Particulars") == ""
    assert parse_period("") == ""


def test_normalize_value():
    assert normalize_value("476,911") == "476911"
    assert normalize_value("2,049,373") == "2049373"
    assert normalize_value("1.008.622") == "1008622"
    assert normalize_value("1234.56") == "1234.56"
    assert normalize_value("(4,721)") == "-4721"
    assert normalize_value("-1,087") == "-1087"
    assert normalize_value("abc") == ""
    assert normalize_value("") == ""


def _rows(*rows):
    return [(i + 1, "", list(cells)) for i, cells in enumerate(rows)]


def test_facts_basic():
    rows = _rows(
        ["Particulars", "31-03-2018", "31-03-2017"],
        ["Assets", "", ""],
        ["Property, plant and equipment", "476,911", "381,176"],
        ["Capital work-in-progress", "27,387", "11,818"],
        ["Total assets", "2,049,373", "1,916,376"],
    )
    facts = list(facts_for_sheet("s.jpg", rows))
    assert len(facts) == 6
    f = facts[0]
    assert f["line_item"] == "Property, plant and equipment"
    assert f["section"] == "Assets"
    assert f["period"] == "2018-03-31"
    assert f["value"] == "476911"
    assert facts[1]["period"] == "2017-03-31"


def test_facts_standalone_consolidated():
    rows = _rows(
        ["", "Standalone", "", "Consolidated", ""],
        ["Particulars", "31-03-2018", "31-03-2017", "31-03-2018", "31-03-2017"],
        ["Inventories", "8,002.02", "13,925.10", "8,002.02", "13,925.10"],
        ["Trade receivables", "18,541.75", "13,671.01", "18,541.75", "13,671.01"],
        ["Cash", "298.82", "741.00", "2,883.82", "477.11"],
    )
    facts = list(facts_for_sheet("s.jpg", rows))
    periods = {f["period"] for f in facts}
    assert "standalone:2018-03-31" in periods
    assert "consolidated:2017-03-31" in periods


def test_wrapped_label_merges():
    rows = _rows(
        ["Particulars", "2018", "2017"],
        ["Investment in subsidiaries and joint", "481,219", "459,538"],
        ["ventures", "", ""],
        ["Inventories", "63", "39"],
        ["Loans", "72,496", "72,081"],
    )
    facts = list(facts_for_sheet("s.jpg", rows))
    assert facts[0]["line_item"] == "Investment in subsidiaries and joint ventures"


def test_pending_label_attaches_to_values_row():
    rows = _rows(
        ["Particulars", "2018", "2017"],
        ["Inventories", "63", "39"],
        ["Loans", "72,496", "72,081"],
        ["Total non-current assets", "", ""],
        ["", "1,834,414", "1,749,546"],
    )
    facts = list(facts_for_sheet("s.jpg", rows))
    assert facts[-2]["line_item"] == "Total non-current assets"
    assert facts[-2]["value"] == "1834414"


def test_no_header_falls_back_to_positional_periods():
    rows = _rows(
        ["Item", "100", "200"],
        ["Item", "100", "200"],
        ["Item", "100", "200"],
    )
    facts = list(facts_for_sheet("s.jpg", rows))
    assert {f["period"] for f in facts} == {"col2", "col3"}
    assert detect_header([r[2] for r in rows]) == (-1, [])
