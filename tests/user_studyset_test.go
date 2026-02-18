package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserStudysets(t *testing.T) {
	// 1. Setup: Create public and private studysets for user1
	// Public Studyset
	createPublicBody := map[string]interface{}{
		"query": `mutation CreateStudyset($studyset: StudysetInput!) {
			createStudyset(studyset: $studyset) {
				id
				title
				private
			}
		}`,
		"variables": map[string]interface{}{
			"studyset": map[string]interface{}{
				"title":     "Public Studyset",
				"private":   false,
				"subjectId": nil,
			},
		},
	}
	req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createPublicBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ := http.DefaultClient.Do(req)
	var createPublicResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createPublicResult)
	require.Nil(t, createPublicResult["errors"])
	publicStudysetID := getNested(createPublicResult, "data", "createStudyset", "id").(string)

	// Private Studyset
	createPrivateBody := map[string]interface{}{
		"query": `mutation CreateStudyset($studyset: StudysetInput!) {
			createStudyset(studyset: $studyset) {
				id
				title
				private
			}
		}`,
		"variables": map[string]interface{}{
			"studyset": map[string]interface{}{
				"title":     "Private Studyset",
				"private":   true,
				"subjectId": nil,
			},
		},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createPrivateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ = http.DefaultClient.Do(req)
	var createPrivateResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createPrivateResult)
	require.Nil(t, createPrivateResult["errors"])
	privateStudysetID := getNested(createPrivateResult, "data", "createStudyset", "id").(string)

	// 2. Test Public Access (Anonymous)
	queryAnon := map[string]interface{}{
		"query": `query GetUserStudysets($userId: ID!) {
			user(id: $userId) {
				studysets {
					edges {
						node {
							id
							title
							private
						}
					}
				}
			}
		}`,
		"variables": map[string]interface{}{
			"userId": user1ID,
		},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(queryAnon))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = http.DefaultClient.Do(req)
	var anonResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&anonResult)
	require.Nil(t, anonResult["errors"])

	edges := getNested(anonResult, "data", "user", "studysets", "edges").([]interface{})
	foundPublic := false
	foundPrivate := false
	for _, edge := range edges {
		node := edge.(map[string]interface{})["node"].(map[string]interface{})
		if node["id"] == publicStudysetID {
			foundPublic = true
		}
		if node["id"] == privateStudysetID {
			foundPrivate = true
		}
	}
	require.True(t, foundPublic, "Anonymous user should see public studyset")
	require.False(t, foundPrivate, "Anonymous user should NOT see private studyset")

	// 3. Test Public Access (Other User - User2)
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(queryAnon))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, _ = http.DefaultClient.Do(req)
	var user2Result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&user2Result)
	require.Nil(t, user2Result["errors"])

	edges = getNested(user2Result, "data", "user", "studysets", "edges").([]interface{})
	foundPublic = false
	foundPrivate = false
	for _, edge := range edges {
		node := edge.(map[string]interface{})["node"].(map[string]interface{})
		if node["id"] == publicStudysetID {
			foundPublic = true
		}
		if node["id"] == privateStudysetID {
			foundPrivate = true
		}
	}
	require.True(t, foundPublic, "Other user should see public studyset")
	require.False(t, foundPrivate, "Other user should NOT see private studyset by default")

	// 4. Test Private Access (Other User - User2 trying to include private)
	queryIncludePrivate := map[string]interface{}{
		"query": `query GetUserStudysets($userId: ID!) {
			user(id: $userId) {
				studysets(includePrivate: true) {
					edges {
						node {
							id
							title
							private
						}
					}
				}
			}
		}`,
		"variables": map[string]interface{}{
			"userId": user1ID,
		},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(queryIncludePrivate))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token) // User2 is NOT mod
	resp, _ = http.DefaultClient.Do(req)
	var user2PrivateResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&user2PrivateResult)
	require.Nil(t, user2PrivateResult["errors"])

	edges = getNested(user2PrivateResult, "data", "user", "studysets", "edges").([]interface{})
	foundPrivate = false
	for _, edge := range edges {
		node := edge.(map[string]interface{})["node"].(map[string]interface{})
		if node["id"] == privateStudysetID {
			foundPrivate = true
		}
	}
	require.False(t, foundPrivate, "Other user (non-mod) should NOT see private studyset even with includePrivate=true")

	// 5. Test Private Access (Owner - User1)
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(queryIncludePrivate))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ = http.DefaultClient.Do(req)
	var user1PrivateResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&user1PrivateResult)
	require.Nil(t, user1PrivateResult["errors"])

	edges = getNested(user1PrivateResult, "data", "user", "studysets", "edges").([]interface{})
	foundPrivate = false
	for _, edge := range edges {
		node := edge.(map[string]interface{})["node"].(map[string]interface{})
		if node["id"] == privateStudysetID {
			foundPrivate = true
		}
	}
	require.True(t, foundPrivate, "Owner should see their own private studyset with includePrivate=true")
}
