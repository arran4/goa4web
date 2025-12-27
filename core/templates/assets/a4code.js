// A4Code Parser and Converter

(function(global) {
    const A4Code = {};

    // --- Tokenizer / Parser helpers ---

    function tokenizeA4Code(text) {
        // Simple tokenizer for [tag args]content[/tag] or [tag args content]
        // Given the spec [tag args content], we need to parse tokens:
        // Text, OpenBracket, CloseBracket, Space
        // Actually, better to parse logically.
        // But for conversion, we might need an AST.
        return parseA4CodeToAST(text);
    }

    // AST Node types: 'root', 'text', 'element'
    // Element: { type: 'element', tagName: string, args: [], children: [] }

    function parseA4CodeToAST(text) {
        const root = { type: 'root', children: [] };
        const stack = [root];
        let current = root;

        let i = 0;
        let len = text.length;

        while (i < len) {
            const char = text[i];

            if (char === '[') {
                // Check if it's an escaped bracket? The Go parser handles `\`
                // Go parser: `\` escapes next char.
                // We should handle that.

                // Peek ahead to see if it's a tag
                // Go parser: `[` triggers acommReader.
                // reads command.

                // Let's implement a lookahead to find the command.
                let j = i + 1;
                let cmdStart = j;
                while (j < len && text[j] !== ' ' && text[j] !== ']' && text[j] !== '[') {
                    j++;
                }
                const cmd = text.substring(cmdStart, j);

                if (cmd.length > 0) {
                     // Start of a tag
                     const node = { type: 'element', tagName: cmd.toLowerCase(), args: [], children: [] };

                     // If there are args before the content starts?
                     // Go parser: `getNextReader` stops at space.
                     // Then `readWhiteSpace`.
                     // Then special handling for specific tags.

                     // img: reads next token as url.
                     // link: reads next token as url.
                     // others: just start content.

                     current.children.push(node);

                     // Handle args based on tag type
                     let k = j;
                     if (node.tagName === 'img' || node.tagName === 'image') {
                         // Read URL
                         while (k < len && text[k] === ' ') k++;
                         let argStart = k;
                         while (k < len && text[k] !== ' ' && text[k] !== ']') k++;
                         node.args.push(text.substring(argStart, k));
                     } else if (node.tagName === 'link' || node.tagName === 'a' || node.tagName === 'url') {
                          // Read URL
                         while (k < len && text[k] === ' ') k++;
                         let argStart = k;
                         while (k < len && text[k] !== ' ' && text[k] !== ']') k++;
                         node.args.push(text.substring(argStart, k));
                     } else if (node.tagName === 'code') {
                         // Code block is special: [code]...[/code]
                         // Go parser: reads until [/code].
                         // It treats everything inside as raw text.
                         // But wait, the Go parser `directOutputReader` handles nesting?
                         // "if buf.Len() >= lens[idx] && ... strings.EqualFold(term, ...)"
                         // It looks for `[/code]` or `code]`.

                         // For 2-way conversion, we should try to support `[code]...[/code]` or `[code ...]`
                         // The Go parser supports `[code]...[/code]` via `directOutputReader`.
                         // So we should search for `[/code]`.

                         let closeTag = '[/code]';
                         let closeIdx = text.toLowerCase().indexOf(closeTag, k);
                         if (closeIdx !== -1) {
                             // Found closing tag
                             let content = text.substring(k + 1, closeIdx); // +1 to skip ']' of [code]?
                             // Wait, we are at 'k' which is after 'code'.
                             // If input is `[code]...`, k is at `]`.
                             if (text[k] === ']') k++;

                             content = text.substring(k, closeIdx);
                             node.children.push({ type: 'text', value: content });

                             i = closeIdx + closeTag.length;
                             continue;
                         }
                     }

                     // consume whitespace after tag/args before content
                     while (k < len && text[k] === ' ') k++;

                     // For normal tags, we push to stack
                     stack.push(node);
                     current = node;
                     i = k;
                     continue;
                } else {
                    // Empty brackets [] or [ ] -> treat as text?
                    // Go parser: empty command -> `stack = append(a.stack, "")`. Pushes empty string to close stack.
                    // Effectively ignores it but consumes `]`.
                    // We'll treat as text for now.
                    current.children.push({ type: 'text', value: '[' });
                    i++;
                }
            } else if (char === ']') {
                if (stack.length > 1) {
                    stack.pop();
                    current = stack[stack.length - 1];
                } else {
                    // Extra closing bracket, treat as text
                    current.children.push({ type: 'text', value: ']' });
                }
                i++;
            } else {
                // Text content
                // Read until next [ or ] or end
                let start = i;
                while (i < len && text[i] !== '[' && text[i] !== ']') {
                    // Handle escapes?
                    if (text[i] === '\\' && i + 1 < len) {
                        i += 2;
                    } else {
                        i++;
                    }
                }
                let val = text.substring(start, i);
                // Unescape?
                current.children.push({ type: 'text', value: val });
            }
        }

        return root;
    }

    // --- AST to Markdown ---

    function astToMarkdown(node) {
        if (node.type === 'root') {
            return node.children.map(astToMarkdown).join('');
        } else if (node.type === 'text') {
            return node.value; // Escaping needed?
        } else if (node.type === 'element') {
            let inner = node.children.map(astToMarkdown).join('');
            switch (node.tagName) {
                case 'b':
                case 'bold':
                    return `**${inner}**`;
                case 'i':
                case 'italic':
                    return `*${inner}*`;
                case 'link':
                case 'a':
                    // [link url text] -> [text](url)
                    // Inner text is the 'text' part.
                    // But wait, `[link url]` works too?
                    // AST args[0] is url.
                    let url = node.args[0] || '';
                    // The inner content is the link text.
                    // Note: In A4Code `[link url text]`, `text` is content.
                    return `[${inner.trim()}](` + url + `)`;
                case 'img':
                case 'image':
                    // [img url] -> ![url](url)
                    // Markdown image: ![alt](src)
                    // A4Code doesn't really support alt text nicely in [img url] style?
                    // Just use url as alt.
                    return `![image](${node.args[0] || ''})`;
                case 'code':
                    return "```\n" + inner + "\n```";
                case 'quote':
                case 'q':
                    // Blockquote
                    return inner.split('\n').map(l => `> ${l}`).join('\n');
                default:
                    return inner;
            }
        }
        return '';
    }

    // --- Markdown to AST (Simplified) ---
    // Implementing a full CommonMark parser is hard.
    // We will support a subset: **bold**, *italic*, [text](url), ![alt](url), `code`, > quote

    function parseMarkdownToAST(text) {
         // This is the hard part "more complex".
         // We can stick to regex-based replacement if we process nesting carefully,
         // or write a scanner.

         // Let's iterate and maintain a stack of open formatting.
         // But Markdown isn't strictly stack-based (e.g. `*bold **bold-italic* bold**`).

         // For the purpose of "2 way convertability" with A4Code (which is stack based),
         // we might assume the Markdown is also well-formed or fix it.

         // Let's try a simple token stream approach.

         const root = { type: 'root', children: [] };
         let current = root;
         let stack = [root]; // Stack of nodes

         // Tokenize by special chars: *, [, !, `, >
         let i = 0;
         while (i < text.length) {
             let char = text[i];

             if (char === '*' && text[i+1] === '*') {
                 // Bold **
                 // Check if closing or opening?
                 // Simple logic: Toggle?
                 // We need to know if we are inside bold.
                 let inBold = stack.some(n => n.type === 'element' && n.tagName === 'b');

                 if (inBold) {
                     // Close bold. Find closest bold in stack.
                     // If it's not top, we have overlapping tags. Close everything up to it.
                     let idx = stack.findLastIndex(n => n.type === 'element' && n.tagName === 'b');
                     if (idx !== -1) {
                         // Pop until idx
                         while (stack.length > idx + 1) {
                             stack.pop();
                         }
                         stack.pop(); // Pop bold
                         current = stack[stack.length-1];
                     }
                 } else {
                     // Open bold
                     let node = { type: 'element', tagName: 'b', children: [] };
                     current.children.push(node);
                     stack.push(node);
                     current = node;
                 }
                 i += 2;
             } else if (char === '*') {
                 // Italic *
                 let inItalic = stack.some(n => n.type === 'element' && n.tagName === 'i');
                 if (inItalic) {
                     let idx = stack.findLastIndex(n => n.type === 'element' && n.tagName === 'i');
                     if (idx !== -1) {
                         while (stack.length > idx + 1) stack.pop();
                         stack.pop();
                         current = stack[stack.length-1];
                     }
                 } else {
                     let node = { type: 'element', tagName: 'i', children: [] };
                     current.children.push(node);
                     stack.push(node);
                     current = node;
                 }
                 i++;
             } else if (char === '`' && text.slice(i, i+3) === '```') {
                 // Code block
                 // Find end
                 i += 3;
                 let end = text.indexOf('```', i);
                 let content;
                 if (end === -1) {
                     content = text.substring(i);
                     i = text.length;
                 } else {
                     content = text.substring(i, end);
                     i = end + 3;
                 }
                 current.children.push({ type: 'element', tagName: 'code', children: [{type:'text', value: content}]});
             } else if (char === '[') {
                 // Link or Image (if ! before)
                 // Wait, logic for image is at '!'

                 // Link: [text](url)
                 // We need to parse [text] first.
                 // This is recursive.

                 // Simpler approach:
                 // Treat [ as start of text, unless followed by ](url) later?
                 // No, [ starts the link text.
                 // We push a 'link' node? But we don't know URL yet.

                 // Let's parse ahead to see if it is a link.
                 // Nested [] not allowed in standard MD links easily.

                 let endBracket = text.indexOf(']', i);
                 if (endBracket !== -1 && text[endBracket+1] === '(') {
                     let endParen = text.indexOf(')', endBracket);
                     if (endParen !== -1) {
                         // It is a link.
                         let linkText = text.substring(i+1, endBracket);
                         let url = text.substring(endBracket+2, endParen);

                         let node = { type: 'element', tagName: 'link', args: [url], children: [] };
                         // Parse inner text?
                         // For now, let's just make it a text node to avoid infinite recursion complexity for this task
                         node.children.push({ type: 'text', value: linkText });

                         current.children.push(node);
                         i = endParen + 1;
                         continue;
                     }
                 }

                 current.children.push({ type: 'text', value: '[' });
                 i++;
             } else if (char === '!' && text[i+1] === '[') {
                 // Image ![alt](url)
                 let endBracket = text.indexOf(']', i);
                 if (endBracket !== -1 && text[endBracket+1] === '(') {
                     let endParen = text.indexOf(')', endBracket);
                     if (endParen !== -1) {
                         let alt = text.substring(i+2, endBracket);
                         let url = text.substring(endBracket+2, endParen);

                         // A4Code image doesn't show text content usually, just [img url]
                         // We will drop alt text or maybe parse it?

                         let node = { type: 'element', tagName: 'img', args: [url], children: [] };
                         current.children.push(node);
                         i = endParen + 1;
                         continue;
                     }
                 }
                 current.children.push({ type: 'text', value: '!' });
                 i++;
             } else {
                 current.children.push({ type: 'text', value: char });
                 i++;
             }
         }

         return root;
    }

    function astToA4Code(node) {
        if (node.type === 'root') {
            return node.children.map(astToA4Code).join('');
        } else if (node.type === 'text') {
            return node.value;
        } else if (node.type === 'element') {
            let inner = node.children.map(astToA4Code).join('');
            switch (node.tagName) {
                case 'b': return `[b ${inner}]`;
                case 'i': return `[i ${inner}]`;
                case 'link':
                    // [link url text]
                    return `[link ${node.args[0]} ${inner}]`;
                case 'img':
                    return `[img ${node.args[0]}]`;
                case 'code':
                    return `[code]${inner}[/code]`;
                default:
                    return inner;
            }
        }
        return '';
    }

    // --- Public API ---

    A4Code.a4codeToMarkdown = function(text) {
        const ast = parseA4CodeToAST(text);
        return astToMarkdown(ast);
    };

    A4Code.markdownToA4Code = function(text) {
        const ast = parseMarkdownToAST(text);
        return astToA4Code(ast);
    };

    global.A4Code = A4Code;

})(this);
