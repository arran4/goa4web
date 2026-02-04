package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

func ManageTopicLabelsPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicID, err := strconv.Atoi(vars["topic"])
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
		return
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	basePath := cd.ForumBasePath
	if basePath == "" {
		if strings.HasPrefix(r.URL.Path, "/private") {
			basePath = "/private"
		} else {
			basePath = "/forum"
		}
	}

	section := "forum"
	if strings.HasPrefix(basePath, "/private") {
		section = "privateforum"
	}

	canLabel, err := UserCanLabelTopic(r.Context(), cd.Queries(), section, int32(topicID), int32(cd.UserID))
	if err != nil {
		log.Printf("UserCanLabelTopic: %v", err)
	}

	if !cd.IsAdmin() && !canLabel {
		handlers.RenderErrorPage(w, r, fmt.Errorf("permission denied"))
		return
	}

	topicRow, err := cd.ForumTopicByID(int32(topicID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
		} else {
			log.Printf("ForumTopicByID Error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		}
		return
	}

	displayTitle := topicRow.Title.String
	if topicRow.Handler == "private" {
		displayTitle = cd.GetPrivateTopicDisplayTitle(topicRow.Idforumtopic, displayTitle)
	}

	var labels []templates.TopicLabel
	if pub, _, err := cd.TopicPublicLabels(topicRow.Idforumtopic); err == nil {
		for _, l := range pub {
			labels = append(labels, templates.TopicLabel{Name: l, Type: "public"})
		}
	} else {
		log.Printf("list public labels: %v", err)
	}

	if priv, err := cd.TopicPrivateLabels(topicRow.Idforumtopic, 0); err == nil {
		for _, l := range priv {
			labels = append(labels, templates.TopicLabel{Name: l, Type: "private"})
		}
	} else {
		log.Printf("list private labels: %v", err)
	}

	sort.Slice(labels, func(i, j int) bool { return labels[i].Name < labels[j].Name })

	data := struct {
		Topic        *ForumtopicPlus
		Labels       []templates.TopicLabel
		BasePath     string
		DisplayTitle string
	}{
		Topic: &ForumtopicPlus{
			Idforumtopic: topicRow.Idforumtopic,
			Title:        topicRow.Title,
		},
		Labels:       labels,
		BasePath:     basePath,
		DisplayTitle: displayTitle,
	}

	ManageTopicLabelsPageTmpl.Handle(w, r, data)
}

const ManageTopicLabelsPageTmpl tasks.Template = "forum/manageTopicLabels.gohtml"
