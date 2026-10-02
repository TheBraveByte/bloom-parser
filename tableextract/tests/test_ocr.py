from tableextract.ocr import clean_cell, is_noise


def test_clean_cell_strips_border_glyphs():
    assert clean_cell("||26,347") == "26,347"
    assert clean_cell("[23,544") == "23,544"
    assert clean_cell("(5,536") == "5,536"
    got = clean_cell("Deferred Tax Liabilities (Net)")
    assert got == "Deferred Tax Liabilities (Net)"
    assert clean_cell("|") == ""


def test_is_noise():
    assert is_noise("|")
    assert is_noise("~-,")
    assert not is_noise("427")
    assert not is_noise("Assets")
