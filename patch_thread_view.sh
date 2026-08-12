#!/bin/bash

# Update threadPage.gohtml to include the forked conversation banner
sed -i '/<div class="label-bar">/i \
    {{ if .Thread.ReplyToThreadId.Valid }}\
        <div class="forked-banner" style="background-color: #f0f8ff; padding: 10px; margin-bottom: 15px; border-left: 4px solid #0066cc;">\
            <strong>Forked Conversation:</strong> This thread is a reply to \
            <a href="{{ $base }}/topic/{{ .Topic.Idforumtopic }}/thread/{{ .Thread.ReplyToThreadId.Int32 }}{{ if .Thread.ReplyToCommentId.Valid }}#c{{ .Thread.ReplyToCommentId.Int32 }}{{ end }}">Thread #{{ .Thread.ReplyToThreadId.Int32 }}</a>.\
        </div>\
    {{ end }}' core/templates/site/domains/forum/threadPage.gohtml
