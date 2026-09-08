const fs = require('fs');
let code = fs.readFileSync('core/templates/assets/pasteimg.js', 'utf8');

const regex = /function isInsideCodeBlock\(textBeforeCaret\) \{[\s\S]*?return inCodeBlock;\n    \}/;

const newFunc = `
    function isInsideCodeBlock(textBeforeCaret) {
        let i = 0;
        let inCodeBlock = false;

        const lowerText = textBeforeCaret.toLowerCase();

        function skipArgPrefix(idx) {
            if (idx >= lowerText.length) return idx;
            const ch = lowerText[idx];
            if (ch === ' ' || ch === '=') {
                idx++;
                if (idx < lowerText.length) {
                    if (lowerText[idx] === '\\n') {
                        idx++;
                    } else if (lowerText[idx] === '\\r') {
                        idx++;
                        if (idx < lowerText.length && lowerText[idx] === '\\n') {
                            idx++;
                        }
                    }
                }
            } else if (ch === '\\n') {
                idx++;
            } else if (ch === '\\r') {
                idx++;
                if (idx < lowerText.length && lowerText[idx] === '\\n') {
                    idx++;
                }
            }
            return idx;
        }

        while (i < lowerText.length) {
            if (!inCodeBlock && !isEscaped(lowerText, i)) {
                if (lowerText.substring(i).startsWith('[codein')) {
                    let nextChar = lowerText[i + 7];
                    if (nextChar === ' ' || nextChar === '\\n' || nextChar === '\\r' || nextChar === ']' || nextChar === '=' || !nextChar) {
                        let j = skipArgPrefix(i + 7);

                        let argFinished = false;
                        if (j < lowerText.length && lowerText[j] === '"') {
                            j++;
                            while (j < lowerText.length) {
                                if (lowerText[j] === '"' && !isEscaped(lowerText, j)) {
                                    j++;
                                    argFinished = true;
                                    break;
                                }
                                j++;
                            }
                        } else {
                            // unquoted GetNextArg uses GetNext(..., false) which stops on newline, ], [, space, \r
                            // AND NOT '=' because endAtEqual is false!
                            while (j < lowerText.length) {
                                if ((lowerText[j] === ' ' || lowerText[j] === ']' || lowerText[j] === '[' || lowerText[j] === '\\n' || lowerText[j] === '\\r') && !isEscaped(lowerText, j)) {
                                    argFinished = true;
                                    break;
                                }
                                j++;
                            }
                            if (j === lowerText.length) {
                                argFinished = false;
                            }
                        }

                        if (!argFinished) {
                            i = lowerText.length;
                            continue;
                        }

                        j = skipArgPrefix(j);

                        inCodeBlock = true;
                        i = j;
                        continue;
                    }
                } else if (lowerText.substring(i).startsWith('[code')) {
                    let nextChar = lowerText[i + 5];
                    if (nextChar === ' ' || nextChar === '\\n' || nextChar === '\\r' || nextChar === ']' || nextChar === '=' || !nextChar) {
                        inCodeBlock = true;
                        let j = skipArgPrefix(i + 5);

                        if (j < lowerText.length && lowerText[j] === ']') {
                            j++;
                        }
                        i = j;
                        continue;
                    }
                }
            }
            if (inCodeBlock && lowerText[i] === ']' && !isEscaped(lowerText, i)) {
                inCodeBlock = false;
            }
            i++;
        }
        return inCodeBlock;
    }
`;

code = code.replace(regex, newFunc.trim());
fs.writeFileSync('core/templates/assets/pasteimg.js', code);
