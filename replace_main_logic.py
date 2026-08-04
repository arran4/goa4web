import re

with open('handlers/forum/forumTopicPage.go', 'r') as f:
    content = f.read()

# Replace buildTopicPageTitle
title_pattern = re.compile(r'\tvar titleParts \[\]string\n\ttitleParts = append\(titleParts, displayTitle\)\n\n\tif topicRow.Handler != "private" \{\n\t\tif data\.Category != nil && data\.Category\.Title\.Valid \{\n\t\t\ttitleParts = append\(titleParts, data\.Category\.Title\.String\)\n\t\t\}\n\t\ttitleParts = append\(titleParts, "Forum"\)\n\t\} else \{\n\t\ttitleParts = append\(titleParts, "Private Forum"\)\n\t\}\n\tcd\.PageTitle = strings\.Join\(titleParts, " - "\)\n')
content = title_pattern.sub('\tcd.PageTitle = buildTopicPageTitle(topicRow, data, displayTitle)\n', content)

# Replace loadTopicThreadsWithLabels
threads_pattern = re.compile(r'\tthreadRows, err := cd\.ForumThreads\(int32\(topicId\)\)\n\tif err != nil && !errors\.Is\(err, sql\.ErrNoRows\) \{\n\t\tlog\.Printf\("Error: ForumThreads: %s", err\)\n\t\thandlers\.RedirectSeeOtherWithError\(w, r, "", err\)\n\t\treturn\n\t\}\n\tthreads := make\(\[\]\*ThreadWithLabels, len\(threadRows\)\)\n\tfor i, r := range threadRows \{\n\t\tt := &ThreadWithLabels\{GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow: r\}\n\t\tvar lbls \[\]templates\.TopicLabel\n\t\tif pub, author, err := cd\.ThreadPublicLabels\(r\.Idforumthread\); err == nil \{\n\t\t\tfor _, l := range pub \{\n\t\t\t\tlbls = append\(lbls, templates\.TopicLabel\{Name: l, Type: "public"\}\)\n\t\t\t\}\n\t\t\tfor _, l := range author \{\n\t\t\t\tlbls = append\(lbls, templates\.TopicLabel\{Name: l, Type: "author"\}\)\n\t\t\t\}\n\t\t\} else \{\n\t\t\tlog\.Printf\("list public labels: %v", err\)\n\t\t\}\n\t\tif priv, err := cd\.ThreadPrivateLabels\(r\.Idforumthread, r\.Firstpostuserid\.Int32\); err == nil \{\n\t\t\tfor _, l := range priv \{\n\t\t\t\tlbls = append\(lbls, templates\.TopicLabel\{Name: l, Type: "private"\}\)\n\t\t\t\tif l == "unread" \{\n\t\t\t\t\tt\.IsUnread = true\n\t\t\t\t\}\n\t\t\t\}\n\t\t\} else \{\n\t\t\tlog\.Printf\("list private labels: %v", err\)\n\t\t\}\n\t\tsort\.Slice\(lbls, func\(i, j int\) bool \{ return lbls\[i\]\.Name < lbls\[j\]\.Name \}\)\n\t\tt\.Labels = lbls\n\t\tthreads\[i\] = t\n\t\}\n\tdata\.Threads = threads\n')

replacement_threads = """\tthreads, err := loadTopicThreadsWithLabels(cd, int32(topicId))
\tif err != nil {
\t\tlog.Printf("Error: ForumThreads: %s", err)
\t\thandlers.RedirectSeeOtherWithError(w, r, "", err)
\t\treturn
\t}
\tdata.Threads = threads
"""
content = threads_pattern.sub(replacement_threads, content)


# Replace loadTopicLabels
labels_pattern = re.compile(r'\tvar labels \[\]templates\.TopicLabel\n\tif pub, _, err := cd\.TopicPublicLabels\(topicRow\.Idforumtopic\); err == nil \{\n\t\tfor _, l := range pub \{\n\t\t\tlabels = append\(labels, templates\.TopicLabel\{Name: l, Type: "public"\}\)\n\t\t\}\n\t\} else \{\n\t\tlog\.Printf\("list public labels: %v", err\)\n\t\}\n\tsort\.Slice\(labels, func\(i, j int\) bool \{ return labels\[i\]\.Name < labels\[j\]\.Name \}\)\n\tdata\.Labels = labels\n')

content = labels_pattern.sub('\tdata.Labels = loadTopicLabels(cd, topicRow.Idforumtopic)\n', content)


with open('handlers/forum/forumTopicPage.go', 'w') as f:
    f.write(content)

print("Main logic replaced successfully")
