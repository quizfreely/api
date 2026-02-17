package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFolderLifecycle(t *testing.T) {
	// 1. Create Folder (Valid Auth - user1)
	createBody := map[string]interface{}{
		"query": `mutation CreateFolder($name: String!) {
			createFolder(name: $name) {
				id
				name
			}
		}`,
		"variables": map[string]interface{}{
			"name": "Test Folder",
		},
	}

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var createResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResult)
	require.NoError(t, err)
	require.Nil(t, createResult["errors"], "should have no errors on creation: %v", createResult["errors"])

	folderID := getNested(createResult, "data", "createFolder", "id").(string)
	require.NotEmpty(t, folderID)
	require.Equal(t, "Test Folder", getNested(createResult, "data", "createFolder", "name"))

	// 2. Rename Folder (Valid Auth - user1)
	renameBody := map[string]interface{}{
		"query": `mutation RenameFolder($id: ID!, $name: String!) {
			renameFolder(id: $id, name: $name) {
				id
				name
			}
		}`,
		"variables": map[string]interface{}{
			"id":   folderID,
			"name": "Renamed Folder",
		},
	}

	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(renameBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var renameResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&renameResult)
	require.NoError(t, err)
	require.Nil(t, renameResult["errors"], "should have no errors on rename")
	require.Equal(t, "Renamed Folder", getNested(renameResult, "data", "renameFolder", "name"))

	// 3. Unauthorized Rename (user2 trying to rename user1's folder)
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(renameBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var unauthorizedRenameResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&unauthorizedRenameResult)
	require.NoError(t, err)
	require.NotNil(t, unauthorizedRenameResult["errors"], "should return authorization error on rename")

	// 4. Query Folder (Valid Auth - user1)
	queryBody := map[string]interface{}{
		"query": `query GetFolder($id: ID!) {
			folder(id: $id) {
				id
				name
			}
		}`,
		"variables": map[string]interface{}{
			"id": folderID,
		},
	}
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(queryBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var queryResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&queryResult)
	require.NoError(t, err)
	require.Nil(t, queryResult["errors"], "should have no errors on query")
	require.Equal(t, "Renamed Folder", getNested(queryResult, "data", "folder", "name"))

	// 5. Query My Folders
	myFoldersBody := map[string]interface{}{
		"query": `query MyFolders {
			myFolders {
				edges {
					node {
						id
						name
					}
				}
			}
		}`,
	}
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(myFoldersBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var myFoldersResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&myFoldersResult)
	require.NoError(t, err)
	require.Nil(t, myFoldersResult["errors"], "should have no errors on myFolders")
	edges := getNested(myFoldersResult, "data", "myFolders", "edges").([]interface{})
	require.True(t, len(edges) > 0, "should have at least one folder")

	// 6. Delete Folder (Valid Auth - user1)
	deleteBody := map[string]interface{}{
		"query": `mutation DeleteFolder($id: ID!) {
			deleteFolder(id: $id)
		}`,
		"variables": map[string]interface{}{
			"id": folderID,
		},
	}

	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(deleteBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var deleteResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&deleteResult)
	require.NoError(t, err)
	require.Nil(t, deleteResult["errors"], "should have no errors on deletion")
	require.Equal(t, folderID, getNested(deleteResult, "data", "deleteFolder").(string))

	// 7. Unauthorized Delete (user2 trying to delete user1's folder)
	// First recreate folder
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	var recreateResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&recreateResult)
	newFolderID := getNested(recreateResult, "data", "createFolder", "id").(string)

	deleteBody["variables"].(map[string]interface{})["id"] = newFolderID
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(deleteBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	var unauthorizedDeleteResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&unauthorizedDeleteResult)
	require.NotNil(t, unauthorizedDeleteResult["errors"], "should return authorization error on delete")

	// Cleanup: Delete the recreated folder
	deleteBody["variables"].(map[string]interface{})["id"] = newFolderID
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(deleteBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	http.DefaultClient.Do(req)
}

func TestStudysetFolderOperations(t *testing.T) {
	// 1. Setup: Create Studyset and Folder (user1)
	createStudysetBody := map[string]interface{}{
		"query": `mutation CreateStudyset($input: StudysetInput!) {
			createStudyset(studyset: $input) { id }
		}`,
		"variables": map[string]interface{}{
			"input": map[string]interface{}{"title": "Folder Test Studyset", "private": false},
		},
	}
	req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createStudysetBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ := http.DefaultClient.Do(req)
	var studysetResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&studysetResult)
	studysetID := getNested(studysetResult, "data", "createStudyset", "id").(string)

	createFolderBody := map[string]interface{}{
		"query": `mutation CreateFolder($name: String!) {
			createFolder(name: $name) { id }
		}`,
		"variables": map[string]interface{}{"name": "Studyset Test Folder"},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createFolderBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ = http.DefaultClient.Do(req)
	var folderResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&folderResult)
	folderID := getNested(folderResult, "data", "createFolder", "id").(string)

	// 2. Set Studyset Folder
	setFolderBody := map[string]interface{}{
		"query": `mutation SetStudysetFolder($studysetId: ID!, $folderId: ID!) {
			setStudysetFolder(studysetId: $studysetId, folderId: $folderId)
		}`,
		"variables": map[string]interface{}{
			"studysetId": studysetID,
			"folderId":   folderID,
		},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(setFolderBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ = http.DefaultClient.Do(req)
	var setResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&setResult)
	require.Nil(t, setResult["errors"], "should set folder successfully")
	require.True(t, getNested(setResult, "data", "setStudysetFolder").(bool))

	// 3. Verify Studyset contains folder info
	queryStudyset := map[string]interface{}{
		"query": `query GetStudyset($id: ID!) {
			studyset(id: $id) {
				folder {
					id
				}
			}
		}`,
		"variables": map[string]interface{}{"id": studysetID},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(queryStudyset))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ = http.DefaultClient.Do(req)
	var queryResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&queryResult)
	require.Equal(t, folderID, getNested(queryResult, "data", "studyset", "folder", "id"))

	// 4. Remove Studyset From Folder
	removeFolderBody := map[string]interface{}{
		"query": `mutation RemoveStudysetFromFolder($studysetId: ID!) {
			removeStudysetFromFolder(studysetId: $studysetId)
		}`,
		"variables": map[string]interface{}{
			"studysetId": studysetID,
		},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(removeFolderBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ = http.DefaultClient.Do(req)
	var removeResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&removeResult)
	require.Nil(t, removeResult["errors"])
	require.True(t, getNested(removeResult, "data", "removeStudysetFromFolder").(bool))

	// 5. Verify Removal
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(queryStudyset))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ = http.DefaultClient.Do(req)
	json.NewDecoder(resp.Body).Decode(&queryResult)
	require.Nil(t, getNested(queryResult, "data", "studyset", "folder"))
}

func TestFolderNoAuth(t *testing.T) {
	createBody := map[string]interface{}{
		"query": `mutation {
			createFolder(name: "No Auth") { id }
		}`,
	}
	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	require.NotNil(t, result["errors"], "should fail without auth")
}

func TestSharedStudysetFolderOperations(t *testing.T) {
	// 1. Setup: Create Public Studyset (user1)
	createStudysetBody := map[string]interface{}{
		"query": `mutation CreateStudyset($input: StudysetInput!) {
			createStudyset(studyset: $input) { id }
		}`,
		"variables": map[string]interface{}{
			"input": map[string]interface{}{"title": "Shared Studyset", "private": false},
		},
	}
	req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createStudysetBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ := http.DefaultClient.Do(req)
	var studysetResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&studysetResult)
	studysetID := getNested(studysetResult, "data", "createStudyset", "id").(string)

	// 2. Setup: Create Folder (user2)
	createFolderBody := map[string]interface{}{
		"query": `mutation CreateFolder($name: String!) {
			createFolder(name: $name) { id }
		}`,
		"variables": map[string]interface{}{"name": "User2 Folder"},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createFolderBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, _ = http.DefaultClient.Do(req)
	var folderResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&folderResult)
	folderID := getNested(folderResult, "data", "createFolder", "id").(string)

	// 3. User2 adds User1's studyset to User2's folder
	setFolderBody := map[string]interface{}{
		"query": `mutation SetStudysetFolder($studysetId: ID!, $folderId: ID!) {
			setStudysetFolder(studysetId: $studysetId, folderId: $folderId)
		}`,
		"variables": map[string]interface{}{
			"studysetId": studysetID,
			"folderId":   folderID,
		},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(setFolderBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, _ = http.DefaultClient.Do(req)
	var setResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&setResult)
	require.Nil(t, setResult["errors"], "user2 should be able to add user1's public studyset to their folder")
	require.True(t, getNested(setResult, "data", "setStudysetFolder").(bool))

	// 4. Verify Studyset is in User2's folder
	// 4. Verify Studyset is in User2's folder
	queryFolder := map[string]interface{}{
		"query": `query GetFolder($id: ID!) {
			folder(id: $id) {
				studysets {
					id
					title
				}
			}
		}`,
		"variables": map[string]interface{}{"id": folderID},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(queryFolder))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, _ = http.DefaultClient.Do(req)
	var queryResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&queryResult)
	require.Nil(t, queryResult["errors"])

	studysets := getNested(queryResult, "data", "folder", "studysets").([]interface{})
	found := false
	for _, s := range studysets {
		sMap := s.(map[string]interface{})
		if sMap["id"] == studysetID {
			found = true
			break
		}
	}
	require.True(t, found, "studyset should be in user2's folder")

	// 5. User2 removes User1's studyset from User2's folder
	removeFolderBody := map[string]interface{}{
		"query": `mutation RemoveStudysetFromFolder($studysetId: ID!) {
			removeStudysetFromFolder(studysetId: $studysetId)
		}`,
		"variables": map[string]interface{}{
			"studysetId": studysetID,
		},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(removeFolderBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, _ = http.DefaultClient.Do(req)
	var removeResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&removeResult)
	require.Nil(t, removeResult["errors"])
	require.True(t, getNested(removeResult, "data", "removeStudysetFromFolder").(bool))

	// 6. Verify Removal
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(queryFolder))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, _ = http.DefaultClient.Do(req)
	json.NewDecoder(resp.Body).Decode(&queryResult)
	studysets = getNested(queryResult, "data", "folder", "studysets").([]interface{})
	found = false
	for _, s := range studysets {
		sMap := s.(map[string]interface{})
		if sMap["id"] == studysetID {
			found = true
			break
		}
	}
	require.False(t, found, "studyset should not be in user2's folder")
}
