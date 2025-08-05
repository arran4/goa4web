package db

import (
	"strings"
	"testing"
)

func TestGlobalGrantsNews(t *testing.T) {
	if !strings.Contains(getNewsPostByIdWithWriterIdAndThreadCommentCount, "g.item_id = s.idsiteNews OR g.item_id IS NULL") {
		t.Errorf("global grant missing in GetNewsPostByIdWithWriterIdAndThreadCommentCount")
	}
	if !strings.Contains(getNewsPostsByIdsForUserWithWriterIdAndThreadCommentCount, "g.item_id = s.idsiteNews OR g.item_id IS NULL") {
		t.Errorf("global grant missing in GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCount")
	}
}

func TestGlobalGrantsFAQ(t *testing.T) {
	if !strings.Contains(getFAQAnsweredQuestions, "g.item_id = faq.idfaq OR g.item_id IS NULL") {
		t.Errorf("global grant missing in GetFAQAnsweredQuestions")
	}
}

func TestGlobalGrantsImageBBS(t *testing.T) {
	if !strings.Contains(listImagePostsByBoardForLister, "g.item_id = i.imageboard_idimageboard OR g.item_id IS NULL") {
		t.Errorf("global grant missing in ListImagePostsByBoardForLister")
	}
	if !strings.Contains(getImagePostByIDForLister, "g.item_id = i.imageboard_idimageboard OR g.item_id IS NULL") {
		t.Errorf("global grant missing in GetImagePostByIDForLister")
	}
}

func TestGlobalGrantsThreads(t *testing.T) {
	if !strings.Contains(getThreadLastPosterAndPerms, "g.item_id = t.idforumtopic OR g.item_id IS NULL") {
		t.Errorf("global grant missing in GetThreadLastPosterAndPerms")
	}
}

func TestGlobalGrantsLinker(t *testing.T) {
	if !strings.Contains(getLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingForUser, "g.item_id = l.idlinker OR g.item_id IS NULL") {
		t.Errorf("global grant missing in GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingForUser")
	}
}
