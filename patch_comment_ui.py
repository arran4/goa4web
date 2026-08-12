import re

with open('core/templates/site/partials/common/comment.gohtml', 'r') as f:
    content = f.read()

content = content.replace('&#x2197;', '[NEW THREAD REPLY]')

replacement = """                        </span>
                            {{ if $.ReplyThreadCounts }}
                            {{ if index $.ReplyThreadCounts $cmt.Idcomments }}
                                {{ $count := index $.ReplyThreadCounts $cmt.Idcomments }}
                                <div style="margin-top: 5px;">
                                    <a href="{{ cd.ForumBasePath }}/topic/{{ $topicID }}/thread/{{ cd.SelectedThreadID }}/replythreads#c{{ $cmt.Idcomments }}">{{ $count }} new thread {{ if eq $count 1 }}reply{{ else }}replies{{ end }}</a>
                                </div>
                            {{ end }}
                            {{ end }}
                    {{ end }}"""

content = re.sub(r'                        </span>\s*{{ end }}', replacement, content)

with open('core/templates/site/partials/common/comment.gohtml', 'w') as f:
    f.write(content)
