#!/bin/bash

sed -i 's/>&#x2197;<\/a>/>[NEW THREAD REPLY]<\/a>/g' core/templates/site/partials/common/comment.gohtml

# We need to insert the counts right after the quote-actions block.
# Let's target the exact closing span of quote-actions.
sed -i '/<\/span><!-- end quote-actions -->/a \
                            {{ if $.ReplyThreadCounts }}\
                            {{ if index $.ReplyThreadCounts $cmt.Idcomments }}\
                                {{ $count := index $.ReplyThreadCounts $cmt.Idcomments }}\
                                <div style="margin-top: 5px;">\
                                    <a href="{{ cd.ForumBasePath }}/topic/{{ $topicID }}/thread/{{ cd.SelectedThreadID }}/replythreads#c{{ $cmt.Idcomments }}">{{ $count }} new thread {{ if eq $count 1 }}reply{{ else }}replies{{ end }}</a>\
                                </div>\
                            {{ end }}\
                            {{ end }}' core/templates/site/partials/common/comment.gohtml
