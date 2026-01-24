import os
import re

directories = ["handlers/forum", "handlers/privateforum"]
import_statement = '\t"github.com/arran4/goa4web/handlers/forumcommon"'

replacements = {
    r'\bRequireThreadAndTopic\b': 'forumcommon.RequireThreadAndTopic',
    r'\bCustomForumIndex\b': 'forumcommon.CustomIndex',
    r'\bBasePathMiddleware\b': 'forumcommon.BasePathMiddleware',
    r'\bSharedTopicPreviewPage\b': 'forumcommon.SharedTopicPreviewPage',
    r'\bSharedThreadPreviewPage\b': 'forumcommon.SharedThreadPreviewPage',
    r'\bQuoteApi\b': 'forumcommon.QuoteApi',
    r'\bsubscribedToTopic\b': 'forumcommon.SubscribedToTopic',
    r'\btopicSubscriptionPattern\b': 'forumcommon.TopicSubscriptionPattern',
    r'\bSectionWeight\b': 'forumcommon.SectionWeight',
}

# Add replacements for Tasks
tasks = [
    "TaskCreateThread", "TaskReply", "TaskEditReply", "TaskCancel", "TaskGrantRole",
    "TaskUpdateRole", "TaskRevokeRole", "TaskSetTopicRestriction", "TaskUpdateTopicRestriction",
    "TaskDeleteTopicRestriction", "TaskCopyTopicRestriction", "TaskRemakeStatisticInformationOnForumthread",
    "TaskRemakeStatisticInformationOnForumtopic", "TaskForumCategoryChange", "TaskForumCategoryCreate",
    "TaskDeleteCategory", "TaskForumThreadDelete", "TaskForumTopicChange", "TaskForumTopicDelete",
    "TaskForumTopicCreate", "TaskTopicGrantCreate", "TaskTopicGrantDelete", "TaskTopicGrantUpdate",
    "TaskCategoryGrantCreate", "TaskCategoryGrantDelete", "TaskSubscribeToTopic", "TaskUnsubscribeFromTopic",
    "TaskAddPublicLabel", "TaskRemovePublicLabel", "TaskAddAuthorLabel", "TaskRemoveAuthorLabel",
    "TaskAddPrivateLabel", "TaskRemovePrivateLabel", "TaskMarkThreadRead", "TaskSetLabels"
]

for task in tasks:
    replacements[r'\b' + task + r'\b'] = 'forumcommon.' + task

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    original_content = content
    modified = False

    # Check if we need to add import
    needs_import = False
    for pattern in replacements.keys():
        if re.search(pattern, content):
            needs_import = True
            break

    # Don't add import if already there
    if "handlers/forumcommon" in content:
        needs_import = False

    # Check if package is forumcommon (don't update self)
    if "package forumcommon" in content:
        return

    if needs_import:
        # Try to insert after other imports
        if "import (" in content:
            content = content.replace("import (", "import (\n" + import_statement, 1)
            modified = True
        elif 'import "' in content:
             # Find the first import and append
             match = re.search(r'import ".*"', content)
             if match:
                 content = content[:match.end()] + "\n" + 'import "github.com/arran4/goa4web/handlers/forumcommon"' + content[match.end():]
                 modified = True

    for pattern, replacement in replacements.items():
        if re.search(pattern, content):
            content = re.sub(pattern, replacement, content)
            modified = True

    if modified:
        with open(filepath, 'w') as f:
            f.write(content)
        print(f"Updated {filepath}")

for d in directories:
    for root, _, files in os.walk(d):
        for file in files:
            if file.endswith(".go"):
                process_file(os.path.join(root, file))
