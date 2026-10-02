from tableextract.normalize import NUMERIC  # noqa: F401 - re-export sanity
from tableextract.refine import normalize_reply, suspect


def test_normalize_reply_unwraps_sentences():
    assert normalize_reply('The text in the cell is "As at".') == "As at"
    assert normalize_reply("The image shows the number 507") == "507"
    assert normalize_reply("**481,219**") == "481,219"
    assert normalize_reply("EMPTY") == ""
    assert normalize_reply("EMPTY.") == ""
    assert normalize_reply("**Transcribed Text:**") == ""
    assert normalize_reply("  Investment property.  ") == "Investment property"


def test_normalize_reply_numeric_column():
    assert normalize_reply("The image shows the number 507", numeric=True) == "507"
    assert normalize_reply("about 1,234.56 lakhs", numeric=True) == "1,234.56"


def test_suspect_guards():
    # too long for the crop
    assert suspect("1,234.56", ")", "10", crop_w=40, numeric_col=True)
    # type mismatch in numeric column
    assert suspect("hello", "x", "10", crop_w=200, numeric_col=True)
    # fragment -> long text
    assert suspect("something long", "an", "20", crop_w=500, numeric_col=False)
    # erasing a confident read
    assert suspect("", "Standalone", "65", crop_w=200, numeric_col=False)
    # unrelated overwrite of confident read
    assert suspect("1,234.56", "Standalone", "65", crop_w=200, numeric_col=False)
    # plausible correction is fine
    assert not suspect("As at", "sat", "10", crop_w=200, numeric_col=False)
    assert not suspect("481,219", "lasa.2a9", "25", crop_w=200, numeric_col=True)
