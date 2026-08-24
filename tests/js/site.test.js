const fs = require('fs');
const path = require('path');
const { TextEncoder } = require('util');

// Mock DOM
global.Node = {
    TEXT_NODE: 3,
    ELEMENT_NODE: 1
};

global.TextEncoder = TextEncoder;

class MockNode {
    constructor(type, content = "", tagName = "") {
        this.nodeType = type;
        this.textContent = content;
        this.tagName = tagName;
        this.childNodes = [];
        this.parentElement = null;
        this.attributes = {};
        this.classList = {
            _classes: new Set(),
            contains: (c) => this.classList._classes.has(c),
            add: (...c) => c.forEach(x => this.classList._classes.add(x)),
            remove: (...c) => c.forEach(x => this.classList._classes.delete(x)),
        };
    }

    get id() { return this.attributes['id'] || ''; }
    set id(val) { this.attributes['id'] = val; }

    get className() { return Array.from(this.classList._classes).join(' '); }
    set className(val) {
        this.classList._classes = new Set(val.split(/\s+/).filter(Boolean));
    }

    appendChild(child) {
        child.parentElement = this;
        this.childNodes.push(child);
    }

    get parentNode() { return this.parentElement; }

    hasAttribute(name) {
        return name in this.attributes;
    }

    getAttribute(name) {
        if (name === 'class') return this.className;
        return this.attributes[name];
    }

    setAttribute(name, value) {
        this.attributes[name] = value;
        if (name === 'class') {
            this.className = value;
        }
    }

    closest(selector) {
        let cur = this;
        while (cur) {
            if (cur.nodeType === Node.ELEMENT_NODE && cur._matches(selector)) {
                return cur;
            }
            cur = cur.parentElement;
        }
        return null;
    }

    _matches(selector) {
        const parts = selector.split(',').map(s => s.trim());
        return parts.some(p => {
            const tokens = p.split(/\s*>\s*|\s+/).filter(Boolean);
            const last = tokens[tokens.length - 1];
            if (last.startsWith('.')) return this.classList.contains(last.substring(1));
            if (last.startsWith('#')) return this.id === last.substring(1);
            if (last.includes('[id^="comment-"]')) {
                return (this.tagName.toLowerCase() === 'div' || !this.tagName) && this.id.startsWith('comment-');
            }
            if (/^[a-zA-Z0-9]+$/.test(last)) {
                return this.tagName.toLowerCase() === last.toLowerCase();
            }
            return false;
        });
    }

    contains(node) {
        if (!node) return false;
        let cur = node;
        while (cur) {
            if (cur === this) return true;
            cur = cur.parentElement;
        }
        return false;
    }
}

global.document = {
    addEventListener: () => {},
    body: {
        addEventListener: () => {}
    }
};
global.window = global;
global.navigator = { clipboard: { writeText: () => Promise.resolve() } };

// Load site.js
const siteJsPath = path.join(__dirname, '../../core/templates/assets/site.js');
const siteJsContent = fs.readFileSync(siteJsPath, 'utf8');
eval(siteJsContent);

// Tests
function assert(condition, message) {
    if (!condition) {
        console.error("FAIL:", message);
        process.exit(1);
    } else {
        console.log("PASS:", message);
    }
}

console.log("Running JS Tests...");

// Test 1: Simple text in a span with start-pos
// <span data-start-pos="10">Hello</span>
// Select "ll" (offset 2). Expected: 10 + 2 = 12.
const span1 = new MockNode(Node.ELEMENT_NODE);
span1.setAttribute('data-start-pos', '10');
const text1 = new MockNode(Node.TEXT_NODE, "Hello");
span1.appendChild(text1);

const res1 = calculateSourceOffset(text1, 2);
assert(res1 === 12, `Simple text offset. Got ${res1}, want 12`);

// Test 2: Multi-byte characters (Emoji)
// <span data-start-pos="100">😀Hello</span>
// Emoji is 4 bytes.
// Select "H" (offset 2 in UTF-16, because 😀 is 2 chars).
// textContent: "\uD83D\uDE00Hello"
// prefix: "😀" (2 chars).
// byteLen of "😀" is 4.
// Expected: 100 + 4 = 104.
const span2 = new MockNode(Node.ELEMENT_NODE);
span2.setAttribute('data-start-pos', '100');
const text2 = new MockNode(Node.TEXT_NODE, "😀Hello");
span2.appendChild(text2);

// Offset of 'H' in "😀Hello" is 2 (surrogate pair counts as 2)
const res2 = calculateSourceOffset(text2, 2);
assert(res2 === 104, `Emoji offset. Got ${res2}, want 104`);

// Test 3: Element selection (offset into childNodes)
// <div data-start-pos="50"><span data-start-pos="60">Child</span></div>
// Select div at offset 0 (before span).
// Should return start pos of child 0 -> 60.
const div3 = new MockNode(Node.ELEMENT_NODE);
div3.setAttribute('data-start-pos', '50');
const span3 = new MockNode(Node.ELEMENT_NODE);
span3.setAttribute('data-start-pos', '60');
div3.appendChild(span3);

const res3 = calculateSourceOffset(div3, 0);
assert(res3 === 60, `Element start offset. Got ${res3}, want 60`);

// Test 4: Fallback to ancestor
// <div data-start-pos="200"><span>Text</span></div>
// Select "Text" offset 0.
// Span has no data-pos. Parent has 200.
// Should return 200.
const div4 = new MockNode(Node.ELEMENT_NODE);
div4.setAttribute('data-start-pos', '200');
const span4 = new MockNode(Node.ELEMENT_NODE); // No pos
div4.appendChild(span4);
const text4 = new MockNode(Node.TEXT_NODE, "Text");
span4.appendChild(text4);

const res4 = calculateSourceOffset(text4, 0);
assert(res4 === 200, `Ancestor fallback. Got ${res4}, want 200`);

// Test 5: End pos
// <div data-end-pos="300"></div>
// Offset 0 (no children).
// Logic: if offset < childNodes.length ... else use end-pos.
// childNodes is empty. 0 is not < 0.
// So returns data-end-pos.
const div5 = new MockNode(Node.ELEMENT_NODE);
div5.setAttribute('data-end-pos', '300');
const res5 = calculateSourceOffset(div5, 0);
assert(res5 === 300, `End pos. Got ${res5}, want 300`);

// Test 6: Text after a rendered <br> inside one annotated span.
// <span data-start-pos="0" data-end-pos="11">Hello<br>World</span>
// Select "Wo" in the second text node. Expected: 0 + len("Hello\nWo") = 8.
const span6 = new MockNode(Node.ELEMENT_NODE, "", "SPAN");
span6.setAttribute('data-start-pos', '0');
span6.setAttribute('data-end-pos', '11');
const text6a = new MockNode(Node.TEXT_NODE, "Hello");
const br6 = new MockNode(Node.ELEMENT_NODE, "", "BR");
const text6b = new MockNode(Node.TEXT_NODE, "World");
span6.appendChild(text6a);
span6.appendChild(br6);
span6.appendChild(text6b);

const res6 = calculateSourceOffset(text6b, 2);
assert(res6 === 8, `Text after br offset. Got ${res6}, want 8`);

// Test 7: Element boundary after a <br> inside one annotated span.
// Boundary before "World" should account for "Hello\n".
const res7 = calculateSourceOffset(span6, 2);
assert(res7 === 6, `Element boundary after br. Got ${res7}, want 6`);

// Test 8: Text after an annotated inline sibling should use the sibling's source length.
// <span data-start-pos="10"><strong data-start-pos="10" data-end-pos="14">Bold</strong> tail</span>
const span8 = new MockNode(Node.ELEMENT_NODE, "", "SPAN");
span8.setAttribute('data-start-pos', '10');
span8.setAttribute('data-end-pos', '19');
const strong8 = new MockNode(Node.ELEMENT_NODE, "", "STRONG");
strong8.setAttribute('data-start-pos', '10');
strong8.setAttribute('data-end-pos', '14');
const strongText8 = new MockNode(Node.TEXT_NODE, "Bold");
const tail8 = new MockNode(Node.TEXT_NODE, " tail");
strong8.appendChild(strongText8);
span8.appendChild(strong8);
span8.appendChild(tail8);

const res8 = calculateSourceOffset(tail8, 3);
assert(res8 === 17, `Text after annotated sibling. Got ${res8}, want 17`);

// Test 9: Text after an annotated image should use the image's source length.
// <span data-start-pos="0"><img data-start-pos="1" data-end-pos="2">b</span>
const span9 = new MockNode(Node.ELEMENT_NODE, "", "SPAN");
span9.setAttribute('data-start-pos', '0');
span9.setAttribute('data-end-pos', '3');
const text9a = new MockNode(Node.TEXT_NODE, "a");
const img9 = new MockNode(Node.ELEMENT_NODE, "", "IMG");
img9.setAttribute('data-start-pos', '1');
img9.setAttribute('data-end-pos', '2');
const text9b = new MockNode(Node.TEXT_NODE, "b");
span9.appendChild(text9a);
span9.appendChild(img9);
span9.appendChild(text9b);

const res9 = calculateSourceOffset(text9b, 1);
assert(res9 === 3, `Text after annotated image. Got ${res9}, want 3`);

const res10 = calculateSourceOffset(span9, 1);
assert(res10 === 1, `Boundary before annotated image. Got ${res10}, want 1`);

const res11 = calculateSourceOffset(span9, 2);
assert(res11 === 2, `Boundary after annotated image. Got ${res11}, want 2`);

console.log("All JS tests passed.");

console.log("Running A4Code Converter Tests...");

const a4codeJsPath_test = path.join(__dirname, '../../core/templates/assets/a4code.js');
let a4codeJsContent_test = fs.readFileSync(a4codeJsPath_test, 'utf8');

a4codeJsContent_test = a4codeJsContent_test.replace('(function(global) {', '');
a4codeJsContent_test = a4codeJsContent_test.replace('})(this);', '');

eval(a4codeJsContent_test);

function assertEqual(actual, expected, msg) {
    if (actual !== expected) {
        console.error(`FAIL: ${msg}`);
        console.error(`Expected:`);
        console.error(JSON.stringify(expected));
        console.error(`Actual:`);
        console.error(JSON.stringify(actual));
        process.exit(1);
    }
}

let md1 = "**bold text**";
let a4_1 = A4Code.markdownToA4Code(md1);
assertEqual(a4_1, "[b bold text]", "Bold markdown to a4code");

let a4_2 = "[b bold text]";
let md2 = A4Code.a4codeToMarkdown(a4_2);
assertEqual(md2, "**bold text**", "Bold a4code to markdown");

let md3 = "Some text\n| col1 | col2 |\n|---|---|\n| val1 | val2 |\nMore text";
let a4_3 = A4Code.markdownToA4Code(md3);
let expected_a4_3 = "Some text\n[code]\n| col1 | col2 |\n|---|---|\n| val1 | val2 |\n[/code]\nMore text";
assertEqual(a4_3, expected_a4_3, "Markdown table wrapped in code block");

let a4_4 = "Some text\n[code]\n| col1 | col2 |\n|---|---|\n| val1 | val2 |\n[/code]\nMore text";
let md4 = A4Code.a4codeToMarkdown(a4_4);
let expected_md4 = "Some text\n\n| col1 | col2 |\n|---|---|\n| val1 | val2 |\n\nMore text";
assertEqual(md4, expected_md4, "A4Code table unwrapped in markdown");

console.log("All Converter Tests Passed!");

let md5 = "# Header 1";
let a4_5 = A4Code.markdownToA4Code(md5);
assertEqual(a4_5, "[h1]Header 1[/h1]", "H1 to a4code");

let a4_6 = "[h2]Header 2[/h2]";
let md6 = A4Code.a4codeToMarkdown(a4_6);
assertEqual(md6, "## Header 2\n", "H2 to md");

let md7 = "> Quote line 1\n> Quote line 2";
let a4_7 = A4Code.markdownToA4Code(md7);
assertEqual(a4_7, "[quote]Quote line 1\nQuote line 2[/quote]", "Quote to a4code");

let a4_8 = "[quote]Quote line 1\nQuote line 2[/quote]";
let md8_res = A4Code.a4codeToMarkdown(a4_8);
assertEqual(md8_res, "> Quote line 1\n> Quote line 2", "Quote to md");

console.log("Running Quote Boundary Selection Tests...");

// Helper to construct a mock comment element tree
function createMockComment(commentId, authorName, commentText) {
    const comment = new MockNode(Node.ELEMENT_NODE, "", "DIV");
    comment.classList.add("comment");
    comment.id = "c" + commentId;

    const aside = new MockNode(Node.ELEMENT_NODE, "", "ASIDE");
    aside.classList.add("author");
    const username = new MockNode(Node.ELEMENT_NODE, "", "DIV");
    username.classList.add("username");
    const usernameText = new MockNode(Node.TEXT_NODE, authorName);
    username.appendChild(usernameText);
    aside.appendChild(username);
    comment.appendChild(aside);

    const section = new MockNode(Node.ELEMENT_NODE, "", "SECTION");
    section.classList.add("body");

    const contentDiv = new MockNode(Node.ELEMENT_NODE, "", "DIV");
    contentDiv.id = "comment-" + commentId;
    const bodyText = new MockNode(Node.TEXT_NODE, commentText);
    contentDiv.appendChild(bodyText);
    section.appendChild(contentDiv);

    const footer = new MockNode(Node.ELEMENT_NODE, "", "FOOTER");
    const quoteActions = new MockNode(Node.ELEMENT_NODE, "", "SPAN");
    quoteActions.classList.add("quote-actions");
    const quoteSelected = new MockNode(Node.ELEMENT_NODE, "", "A");
    quoteSelected.classList.add("quote-link");
    quoteSelected.setAttribute("data-quote-type", "selected");
    quoteSelected.setAttribute("data-comment-id", String(commentId));
    const quoteText = new MockNode(Node.TEXT_NODE, "QUOTE SELECTED");
    quoteSelected.appendChild(quoteText);
    quoteActions.appendChild(quoteSelected);
    footer.appendChild(quoteActions);
    section.appendChild(footer);

    comment.appendChild(section);

    return {
        comment,
        aside,
        usernameText,
        section,
        contentDiv,
        bodyText,
        footer,
        quoteText
    };
}

const c1 = createMockComment(101, "Alice", "Hello world from comment 101");
const c2 = createMockComment(102, "Bob", "Another comment 102");

// Test getCommentContentElement
assert(getCommentContentElement(c1.bodyText) === c1.contentDiv, "getCommentContentElement on text inside comment content");
assert(getCommentContentElement(c1.contentDiv) === c1.contentDiv, "getCommentContentElement on content div directly");
assert(getCommentContentElement(c1.usernameText) === null, "getCommentContentElement on author username should be null");
assert(getCommentContentElement(c1.aside) === null, "getCommentContentElement on author aside should be null");
assert(getCommentContentElement(c1.quoteText) === null, "getCommentContentElement on footer actions should be null");
assert(getCommentContentElement(c1.footer) === null, "getCommentContentElement on footer should be null");

// Test getValidSelectedComment
// 1. Valid selection wholly inside c1 content
const validSelection = {
    isCollapsed: false,
    rangeCount: 1,
    getRangeAt: () => ({
        startContainer: c1.bodyText,
        startOffset: 0,
        endContainer: c1.bodyText,
        endOffset: 5,
        commonAncestorContainer: c1.bodyText
    })
};
const validResult = getValidSelectedComment(validSelection);
assert(validResult !== null, "Valid selection should return result");
assert(validResult.comment === c1.comment, "Valid selection should identify c1.comment");
assert(validResult.contentElement === c1.contentDiv, "Valid selection should identify c1.contentDiv");

// 2. Collapsed selection
const collapsedSelection = {
    isCollapsed: true,
    rangeCount: 1,
    getRangeAt: () => ({
        startContainer: c1.bodyText,
        startOffset: 0,
        endContainer: c1.bodyText,
        endOffset: 0,
        commonAncestorContainer: c1.bodyText
    })
};
assert(getValidSelectedComment(collapsedSelection) === null, "Collapsed selection should return null");

// 3. Selection starting in author metadata and ending in comment content
const authorToContentSelection = {
    isCollapsed: false,
    rangeCount: 1,
    getRangeAt: () => ({
        startContainer: c1.usernameText,
        startOffset: 0,
        endContainer: c1.bodyText,
        endOffset: 5,
        commonAncestorContainer: c1.comment
    })
};
assert(getValidSelectedComment(authorToContentSelection) === null, "Author to content selection should return null");

// 4. Selection starting in comment content and ending in footer
const contentToFooterSelection = {
    isCollapsed: false,
    rangeCount: 1,
    getRangeAt: () => ({
        startContainer: c1.bodyText,
        startOffset: 0,
        endContainer: c1.quoteText,
        endOffset: 3,
        commonAncestorContainer: c1.section
    })
};
assert(getValidSelectedComment(contentToFooterSelection) === null, "Content to footer selection should return null");

// 5. Selection crossing between two different comments
const crossCommentSelection = {
    isCollapsed: false,
    rangeCount: 1,
    getRangeAt: () => ({
        startContainer: c1.bodyText,
        startOffset: 0,
        endContainer: c2.bodyText,
        endOffset: 5,
        commonAncestorContainer: new MockNode(Node.ELEMENT_NODE, "", "BODY")
    })
};
assert(getValidSelectedComment(crossCommentSelection) === null, "Cross-comment selection should return null");

console.log("All Quote Boundary Selection Tests Passed!");

console.log("Running Forum Filter JS Tests...");

const forumJsPath = path.join(__dirname, '../../handlers/forum/forum.js');
const forumJsContent = fs.readFileSync(forumJsPath, 'utf8');

// Extract the parser and evaluate functions from forum.js
let forumScope = {};
(function() {
    const fnStr = forumJsContent
        .replace(/document\.addEventListener\('DOMContentLoaded',\s*\(\)\s*=>\s*\{/, '(function() {')
        .replace(/const labelFilter = document\.querySelector\('\.label-filter'\);[\s\S]*$/, 'return { tokenize, parse, evaluateAST }; })()');
    forumScope = eval(fnStr);
})();

const { tokenize, parse, evaluateAST } = forumScope;

// Test forum filtering across labels, participants, posters, topics
const labels = ["announcement", "help"];
const participants = ["alice", "bob"];
const posters = ["charlie"];
const topics = ["general discussion"];

function testFilterMatch(query, expected, desc) {
    const tokens = tokenize(query);
    const ast = parse(tokens);
    const matched = evaluateAST(ast, labels, participants, posters, topics);
    assert(matched === expected, `${desc} (query: "${query}"). Got ${matched}, want ${expected}`);
}

testFilterMatch("label:announcement", true, "Match existing label");
testFilterMatch("label:bug", false, "Reject non-existing label");
testFilterMatch("poster:charlie", true, "Match existing poster");
testFilterMatch("poster:alice", false, "Reject non-existing poster");
testFilterMatch("participant:alice", true, "Match existing participant");
testFilterMatch("participant:david", false, "Reject non-existing participant");
testFilterMatch("topic:\"general discussion\"", true, "Match existing topic with quotes");
testFilterMatch("topic:rules", false, "Reject non-existing topic");
testFilterMatch("poster:charlie topic:general", true, "Match implicit AND");
testFilterMatch("poster:charlie (label:help OR label:announcement)", true, "Match compound AND with OR");
testFilterMatch("poster:alice OR poster:charlie", true, "Match compound OR");
testFilterMatch("poster:alice OR topic:missing", false, "Reject failed OR");

console.log("All Forum Filter JS Tests Passed!");

