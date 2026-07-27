#!/bin/bash
cat << 'TEMPLATE_EOF' > core/templates/site/privateforum/unread.gohtml
{{ define "privateforum/unread.gohtml" }}
    {{ template "head" $ }}
<link rel="stylesheet" href="{{ assetHash "/forum/forum.css" }}">
<script src="{{ assetHash "/forum/forum.js" }}"></script>
    <h1>Unread Private Threads</h1>
    {{ if .CurrentError }}
        <p class="error-message">{{ .CurrentError }}</p>
    {{ end }}
    {{ if eq (len .Threads) 0 }}
        <p>You have no unread private threads.</p>
    {{ else }}
        {{ $base := "/private" }}
        <div class="thread-controls">
            <input type="text" class="label-filter" placeholder="Filter by label...">
            <div class="sort-buttons">
                <span>Sort by:</span>
                <button class="sort-button" data-sort="last-reply" data-order="desc">Last Reply</button>
                <button class="sort-button" data-sort="creation" data-order="desc">Creation</button>
                <button class="sort-button" data-sort="comments" data-order="desc">Comments</button>
            </div>
        </div>
        <div class="thread-list">
        {{- range .Threads }}
            <div class="thread" data-last-reply-time="{{ .Lastaddition.Time.Unix }}" data-creation-time="{{ .Firstpostwritten.Time.Unix }}" data-comments="{{ .Comments.Int32 }}">
                <div class="thread-meta thread-header">
                    First poster: <span class="poster-name first" {{ if ne .Firstpostuserid.Int32 $.cd.UserID }}style="font-weight: bold;"{{ end }}>{{ .Firstpostusername.String }}</span>
                    At <span class="post-time first">{{ $.cd.LocalTime .Firstpostwritten.Time }}</span>
                    in Topic: <strong>{{ .DisplayTitle | default "Untitled Topic" }}</strong>
                </div>
                <div class="thread-content foldable">
                    {{ .Firstposttext.String | a4code2html }}<br>
                </div>
                <div class="thread-meta">
                    Lastposter: <span class="poster-name last" {{ if ne .Lastposterid.Int32 $.cd.UserID }}style="font-weight: bold;"{{ end }}>{{ .Lastposterusername.String }}</span>
                    At <span class="post-time last">{{ $.cd.LocalTime .Lastaddition.Time }}</span>
                    [<a href="{{$base}}/topic/{{.TopicID}}/thread/{{ .Idforumthread }}">
                        {{- .Comments.Int32 }} comments.</a>]{{" "}}
                </div>
            </div>
        {{- end }}
        </div>
        <div class="pagination">
            {{ if gt .Page 1 }}
                <a href="/private/unread?page={{ .PrevPage }}">Previous Page</a>
            {{ end }}
            {{ if and (gt .Page 1) .HasNextPage }} | {{ end }}
            {{ if .HasNextPage }}
                <a href="/private/unread?page={{ .NextPage }}">Next Page</a>
            {{ end }}
        </div>
    {{ end }}
    {{ template "tail" $ }}
{{ end }}
TEMPLATE_EOF
