package graph

import (
	"encoding/base64"
	"quizfreely/api/graph/model"
	"strings"
)

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// SavedStudysetRow is used for MySavedStudysets to include saved_at for cursor.
type SavedStudysetRow struct {
	model.Studyset
	SavedAt *string `db:"saved_at"`
}

// EncodeStudysetCursor encodes (timestampStr, id) for keyset pagination.
// timestampStr is the sort key (e.g. created_at or updated_at from to_char).
func EncodeStudysetCursor(timestampStr, id string) string {
	if timestampStr == "" {
		timestampStr = "|"
	}
	return base64.StdEncoding.EncodeToString([]byte(timestampStr + "|" + id))
}

// StudysetConnectionFrom builds a StudysetConnection from nodes.
// cursorTS and cursorID are the sort key and id of the "after" cursor (for hasPreviousPage).
// getCursor returns (timestampStr, id) for each node to encode as cursor.
func StudysetConnectionFrom(
	nodes []*model.Studyset,
	hasNext bool,
	hasPrevious bool,
	getCursor func(*model.Studyset) (timestampStr, id string),
) *model.StudysetConnection {
	edges := make([]*model.StudysetEdge, 0, len(nodes))
	for _, n := range nodes {
		ts, id := getCursor(n)
		edges = append(edges, &model.StudysetEdge{
			Node:   n,
			Cursor: EncodeStudysetCursor(ts, id),
		})
	}
	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}
	return &model.StudysetConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}
}

// DecodeStudysetCursor returns (timestampStr, id). Empty strings if invalid.
func DecodeStudysetCursor(cursor string) (timestampStr, id string) {
	if cursor == "" {
		return "", ""
	}
	raw, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return "", ""
	}
	s := string(raw)
	parts := strings.Split(s, "|")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

// EncodeStudysetScoreCursor encodes (score, timestampStr, id) for search pagination.
// score is the similarity score (float string).
func EncodeStudysetScoreCursor(score, timestampStr, id string) string {
	if timestampStr == "" {
		timestampStr = ""
	}
	return base64.StdEncoding.EncodeToString([]byte(score + "|" + timestampStr + "|" + id))
}

// DecodeStudysetScoreCursor returns (score, timestampStr, id).
func DecodeStudysetScoreCursor(cursor string) (score, timestampStr, id string) {
	if cursor == "" {
		return "", "", ""
	}
	raw, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return "", "", ""
	}
	s := string(raw)
	parts := strings.Split(s, "|")
	if len(parts) != 3 {
		return "", "", ""
	}
	return parts[0], parts[1], parts[2]
}

// EncodeFolderCursor encodes folder id for keyset pagination (order by id).
func EncodeFolderCursor(id string) string {
	return base64.StdEncoding.EncodeToString([]byte("folder|" + id))
}

// DecodeFolderCursor returns folder id, or "" if invalid.
func DecodeFolderCursor(cursor string) string {
	if cursor == "" {
		return ""
	}
	raw, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return ""
	}
	s := string(raw)
	if !strings.HasPrefix(s, "folder|") {
		return ""
	}
	return s[7:]
}
