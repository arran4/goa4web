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
        <div class="thread-list">
            <ul>
                {{ range .Threads }}
                <li class="thread-item">
                    <a href="/private/topic/{{ .TopicID }}/thread/{{ .Idforumthread }}">{{ .DisplayTitle | default "Untitled Topic" | a4code2html }} ({{ (cd.LocalTime .Lastaddition.Time).Format "2006-01-02 15:04" }})</a> - Last post by {{ .Lastposterusername.String | default "Unknown" }}
                </li>
                {{ end }}
            </ul>
        </div>
        <div class="pagination">
            {{ if gt .Page 1 }}
                <a href="/private/unread?page={{ add .Page -1 }}">Previous Page</a>
            {{ end }}
            {{ if and (gt .Page 1) .HasNextPage }} | {{ end }}
            {{ if .HasNextPage }}
                <a href="/private/unread?page={{ add .Page 1 }}">Next Page</a>
            {{ end }}
        </div>
    {{ end }}
    {{ template "tail" $ }}
{{ end }}
TEMPLATE_EOF
