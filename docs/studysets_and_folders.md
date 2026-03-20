# Study Sets and Folders

Study sets are the core of the user experience. They store collections of terms and can be organized into folders.

## Core Entities

### Study Sets
Individual study sets are stored in the `public.studysets` table.
- **ID:** A unique UUID for the study set.
- **Title:** The name of the study set.
- **Privacy:** A `private` boolean flag. Private sets are only visible to the user who created them.
- **Draft Status:** A `draft` boolean flag. Drafts can be excluded from public search and views.
- **User ID:** Links the study set to the user who created it (`auth.users`).
- **Subject ID:** References a category in the `public.subjects` table.
- **Timestamps:** Tracks creation and last update.

### Folders
Folders help users categorize their study sets. They are stored in the `public.folders` table.
- **ID:** A unique UUID.
- **Name:** The user-defined folder name.
- **User ID:** Links the folder to the user who created it (`auth.users`).

### Relationships (Folders and Study Sets)
The relationship between study sets and folders is managed by the `public.folder_studysets` join table.
- **Folder ID:** References a folder.
- **Studyset ID:** References a study set.
- **User ID:** Ensures each folder-set mapping is specific to a user.

## Key Operations (GraphQL)

### Study Sets
Study sets are created and modified using GraphQL mutations:
- `createStudyset(studyset: StudysetInput!, draft: Boolean!, folderId: ID): Studyset`
- `updateStudyset(id: ID!, studyset: StudysetInput, draft: Boolean!): Studyset`
- `deleteStudyset(id: ID!): ID`

Study sets can be retrieved via various queries:
- `studyset(id: ID!): Studyset`
- `studysets(cursor: String, limit: Int): StudysetConnection` (supports pagination)

### Folders
Folders are managed using dedicated mutations:
- `createFolder(name: String!): Folder`
- `renameFolder(id: ID!, name: String!): Folder`
- `deleteFolder(id: ID!): ID`

A study set can be added to or removed from a folder:
- `setStudysetFolder(studysetId: ID!, folderId: ID!): Boolean`
- `removeStudysetFromFolder(studysetId: ID!): Boolean`

## Saving Study Sets
Users can "save" study sets created by others, which adds them to their collection. This is tracked in the `public.saved_studysets` table and managed with:
- `saveStudyset(studysetId: ID!): Boolean`
- `unsaveStudyset(studysetId: ID!): Boolean`
