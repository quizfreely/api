# Terms and Images

Terms are the fundamental building blocks of study sets. They can contain text and images on both sides (term and definition).

## Core Entities

### Terms
Each term in a study set is stored in the `public.terms` table.
- **ID:** A unique UUID.
- **Term:** The text for the "term" side.
- **Def:** The text for the "definition" side.
- **Sort Order:** An integer used to determine the display order within a study set.
- **Studyset ID:** References the `public.studysets` table.
- **Timestamps:** Tracks creation and last update.

### Images
Images associated with terms are managed through a combination of cloud storage (S3) and a database record in the `public.images` table.
- **Object Key:** A unique identifier for the image file (e.g., `images/ab/cd/ef...webp`).

The `public.terms` table has two fields to link terms to images:
- **Term Image Key:** References an image from `public.images` for the "term" side.
- **Def Image Key:** References an image for the "definition" side.

## Key Operations

### Managing Terms (GraphQL)
Terms are primarily managed via the GraphQL API:
- **Create:** `createTerms(studysetId: ID!, terms: [NewTermInput!]!): [Term]`
- **Update:** `updateTerms(studysetId: ID!, terms: [TermInput!]!): [Term!]`
- **Delete:** `deleteTerms(studysetId: ID!, ids: [ID!]!): [ID!]`

These mutations allow for bulk operations within a single study set.

### Image Uploads (REST)
Image management is handled through dedicated REST endpoints to simplify file processing and S3 integration:
- `PUT /term-images/{termID}/{side}`: Handles image uploading.
- `DELETE /term-images/{termID}/{side}`: Removes the image association from a term.

The `side` parameter can be either `term` or `def`.

#### Image Upload Process:
1. **Validation:** The server checks if the user owns the study set containing the term.
2. **Processing:** The uploaded image is resized and converted to WebP format for optimized delivery.
3. **Storage:** The processed image is uploaded to S3.
4. **Database Update:** A record is added to `public.images` (if it doesn't already exist), and the `term_image_key` or `def_image_key` in the `public.terms` table is updated.

The full image URL is then returned, constructed by prefixing the `UsercontentBaseURL` to the `objectKey`.
