from playwright.sync_api import sync_playwright, expect
import os

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    page = browser.new_page()

    # Load the local HTML file
    file_path = os.path.abspath("verification/test.html")
    page.goto(f"file://{file_path}")

    # 1. Test Full Reply
    print("Testing Full Reply...")
    page.click("a[data-quote-type='full']")
    # Check if fetch was called (by checking window.fetchCalls)
    fetch_calls = page.evaluate("window.fetchCalls")
    print(f"Fetch calls after full reply: {fetch_calls}")
    assert "/api/forum/quote/123?type=full" in fetch_calls

    # Check if text was appended
    reply_val = page.evaluate("document.getElementById('reply').value")
    assert "> Quoted text" in reply_val

    # 2. Test Paragraph Reply
    print("Testing Paragraph Reply...")
    page.click("a[data-quote-type='paragraph']")
    fetch_calls = page.evaluate("window.fetchCalls")
    print(f"Fetch calls after paragraph reply: {fetch_calls}")
    assert "/api/forum/quote/123?type=paragraph" in fetch_calls

    # 3. Test Selected Reply (Mocking selection first)
    print("Testing Selected Reply...")
    page.evaluate("""
        const range = document.createRange();
        range.selectNodeContents(document.body.querySelector('.comment'));
        const sel = window.getSelection();
        sel.removeAllRanges();
        sel.addRange(range);
    """)
    page.click("a[data-quote-type='selected']")
    fetch_calls = page.evaluate("window.fetchCalls")
    print(f"Fetch calls after selected reply: {fetch_calls}")
    # The start/end might vary depending on whitespace, just check the prefix
    assert any("type=selected" in call for call in fetch_calls)

    # Take screenshot
    page.screenshot(path="verification/quote_test.png")

    browser.close()

with sync_playwright() as playwright:
    run(playwright)
