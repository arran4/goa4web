package main

import (
	"fmt"
	"strings"
)

func anchorLink(anchor, linkName string) string {
	return fmt.Sprintf("<a href=\"#%s\">%s</a><br>", anchor, linkName)
}

func pageLink(pagename, linkName string) string {
	return fmt.Sprintf("<a href=\"?page=%s\">%s</a><br>", pagename, linkName)
}

func categoryLevel(categoryName string, level int) string {
	span := ""
	switch level {
	case 0:
		span = "font-size: 20;"
	case 1:
		span = "font-size: 16;"
	case 2:
		span = "font-size: 14;"
	}
	return fmt.Sprintf("<p><a name=\"%s\"><span style=\"%s\">%s</span></a><br>\n", categoryName, span, categoryName)
}

func externalLink(pagename, linkName string) string {
	return fmt.Sprintf("<a href=\"%s\" target=\"_blank\">%s</a>", pagename, linkName)
}

func formatCategories(input string) string {
	var (
		formatter          strings.Builder
		currentLevel       uint
		nlcount            uint
		categoryname       strings.Builder
		sortedListToggle   bool
		unsortedListToggle bool
	)

	formatter.WriteString("<ul>\n")

	for i := 0; i < len(input); i++ {
		nlcount++
		switch input[i] {
		case '\n':
			nlcount = 0
			if sortedListToggle && input[i+1] != '#' {
				sortedListToggle = false
				formatter.WriteString("</ol>\n")
			}
			if unsortedListToggle && input[i+1] != '-' {
				unsortedListToggle = false
				formatter.WriteString("</ul>\n")
			}
			if input[i+1] == '\n' {
				formatter.WriteString("\n<p>")
				i++
			} else {
				formatter.WriteString("<br>\n")
			}
			for i+1 < len(input) && input[i+1] == '\n' {
				i++
			}
		case '=':
			if nlcount == 1 && input[i+1] == '=' {
				level := 0
				for ; i+level+2 < len(input) && input[i+level+2] == '='; level++ {
				}
				for i+level+2 < len(input) && input[i+level+2] != '\n' &&
					!(input[i+level+2] == '=' && input[i+level+3] == '=' &&
						(input[i+level+4] == '\n' || input[i+level+4] == 0)) {
					categoryname.WriteByte(input[i+level+2])
					level++
				}
				if input[i+level+2] != '=' {
					break
				}
				formatter.WriteString(anchorLink(categoryname.String(), categoryname.String()))
				i = i + level + 2 + func() int {
					if input[i+level+3] == '\n' {
						return 1
					}
					return 0
				}()
				nlcount = 0
			}
		}
	}

	formatter.WriteString("</ul>\n")
	for currentLevel > 0 {
		formatter.WriteString("</ul>\n")
		currentLevel--
	}

	return formatter.String()
}

func formatBlob(input string) string {
	var (
		formatter          strings.Builder
		strongToggle       bool
		uToggle            bool
		iToggle            bool
		formatToggle       bool
		unsortedListToggle bool
		sortedListToggle   bool
		htmlToggle         bool
		nlcount            uint
	)

	for i := 0; i < len(input); i++ {
		nlcount++
		if !formatToggle {
			if input[i] == '\\' && i+7 < len(input) &&
				input[i+1] == '<' &&
				input[i+2] == '/' &&
				input[i+3] == 'p' &&
				input[i+4] == 'r' &&
				input[i+5] == 'e' &&
				input[i+6] == '\\' &&
				input[i+7] == '>' {
				formatToggle = true
				formatter.WriteString("</pre>\n")
				i += 7
			} else {
				formatter.WriteByte(input[i])
			}
		} else {
			switch input[i] {
			case '\n':
				nlcount = 0
				if sortedListToggle && input[i+1] != '#' {
					sortedListToggle = false
					formatter.WriteString("</ol>\n")
				}
				if unsortedListToggle && input[i+1] != '-' {
					unsortedListToggle = false
					formatter.WriteString("</ul>\n")
				}
				if input[i+1] == '\n' {
					formatter.WriteString("\n<p>")
					i++
				} else {
					formatter.WriteString("<br>\n")
				}
				for i+1 < len(input) && input[i+1] == '\n' {
					i++
				}
			case '&':
				formatter.WriteString("&amp;")
			case '<':
				if htmlToggle {
					formatter.WriteString("&lt;")
				} else {
					formatter.WriteByte(input[i])
				}
			case '>':
				if htmlToggle {
					formatter.WriteString("&gt;")
				} else {
					formatter.WriteByte(input[i])
				}
			case '#':
				if nlcount == 1 {
					if sortedListToggle {
						formatter.WriteString("<li>")
					} else {
						formatter.WriteString("<ol>\n<li>")
					}
					sortedListToggle = true
				} else {
					formatter.WriteByte('#')
				}
			case '-':
				if nlcount == 1 {
					if unsortedListToggle {
						formatter.WriteString("<li>")
					} else {
						formatter.WriteString("<ul>\n<li>")
					}
					unsortedListToggle = true
				} else {
					formatter.WriteByte('-')
				}
			case '\\':
				if i+1 < len(input) {
					switch input[i+1] {
					case '<':
						if i+7 < len(input) &&
							input[i+2] == 'p' &&
							input[i+3] == 'r' &&
							input[i+4] == 'e' &&
							input[i+5] == '\\' &&
							input[i+6] == '>' {
							formatToggle = false
							formatter.WriteString("<pre>")
							i += 6
						}
					case '*':
						strongToggle = !strongToggle
						if strongToggle {
							formatter.WriteString("<strong>")
						} else {
							formatter.WriteString("</strong>")
						}
						i++
					case '_':
						uToggle = !uToggle
						if uToggle {
							formatter.WriteString("<u>")
						} else {
							formatter.WriteString("</u>")
						}
						i++
					case '/':
						iToggle = !iToggle
						if iToggle {
							formatter.WriteString("<i>")
						} else {
							formatter.WriteString("</i>")
						}
						i++
					case 'h':
						htmlToggle = !htmlToggle
						i++
					case 0:
						i++
					default:
						formatter.WriteByte(input[i])
					}
				}
			case '[':
				if i+1 < len(input) && input[i+1] == '[' {
					var (
						linkaddr strings.Builder
						linkname strings.Builder
					)
					for j := i + 2; j < len(input) && input[j] != '\n' &&
						(input[j] != '|' || (input[j] == '|' && input[j+1] == '|')) &&
						!(input[j] == ']' && input[j+1] == ']'); j++ {
						linkaddr.WriteByte(input[j])
					}
					if i+1 < len(input) && input[i+2] == '|' {
						for j := i + 3; j < len(input) && input[j] != '\n' &&
							!(input[j] == ']' && input[j+1] == ']'); j++ {
							linkname.WriteByte(input[j])
						}
					} else {
						linkname = linkaddr
					}
					if i+1 < len(input) && input[i+2] == '|' {
						i++
					}
					formatter.WriteString(externalLink(linkaddr.String(), linkname.String()))
					i++
					break
				}
				formatter.WriteByte(input[i])
			default:
				formatter.WriteByte(input[i])
			}
		}
	}
	return formatter.String()
}
