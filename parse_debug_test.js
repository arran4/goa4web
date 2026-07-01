const { TextEncoder } = require('util');

const Node = {
    TEXT_NODE: 3,
    ELEMENT_NODE: 1
};

global.TextEncoder = TextEncoder;

class MockNode {
    constructor(type, content = "") {
        this.nodeType = type;
        this.textContent = content;
        this.childNodes = [];
        this.parentElement = null;
        this.attributes = {};
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
        return this.attributes[name];
    }

    setAttribute(name, value) {
        this.attributes[name] = value;
    }
}

function calculateSourceOffset(node, offset) {
    if (node.nodeType === Node.TEXT_NODE) {
        // Look for parent with data-start-pos or data-comment-offset
        const parent = node.parentElement;
        if (parent) {
            let startAttr = parent.getAttribute('data-comment-offset') || parent.getAttribute('data-start-pos');
            if (startAttr !== null) {
                const baseStart = parseInt(startAttr, 10);
                const textContent = node.textContent;
                const prefix = textContent.substring(0, offset);
                const byteLen = new TextEncoder().encode(prefix).length;
                return baseStart + byteLen;
            }
        }
    } else if (node.nodeType === Node.ELEMENT_NODE) {
        // If offset points to a child, try to find start pos of that child
        if (offset < node.childNodes.length) {
            const child = node.childNodes[offset];
            if (child.nodeType === Node.ELEMENT_NODE) {
                let startAttr = child.getAttribute('data-comment-offset') || child.getAttribute('data-start-pos');
                if (startAttr !== null) {
                    return parseInt(startAttr, 10);
                }
            } else if (child.nodeType === Node.TEXT_NODE) {
                 return calculateSourceOffset(child, 0);
            }
        } else {
             // Offset at end.
             if (node.hasAttribute('data-end-pos')) {
                 return parseInt(node.getAttribute('data-end-pos'), 10);
             }
        }
    }
    // Fallback: try to find nearest ancestor with data-start-pos
    let current = node;
    while (current) {
        if (current.nodeType === Node.ELEMENT_NODE) {
            let startAttr = current.getAttribute('data-comment-offset') || current.getAttribute('data-start-pos');
            if (startAttr !== null && startAttr !== undefined) {
                return parseInt(startAttr, 10);
            }
        }
        current = current.parentNode;
    }
    return -1;
}

const div4 = new MockNode(Node.ELEMENT_NODE);
div4.setAttribute('data-start-pos', '200');
const span4 = new MockNode(Node.ELEMENT_NODE); // No pos
div4.appendChild(span4);
const text4 = new MockNode(Node.TEXT_NODE, "Text");
span4.appendChild(text4);

console.log(calculateSourceOffset(text4, 0));
