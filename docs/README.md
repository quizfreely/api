# Quizfreely API Documentation

This directory contains documentation for the Quizfreely API, outlining how core entities are stored, updated, and accessed via both GraphQL and REST.

## Navigation

- [**Authentication and User Management**](./authentication.md): Covers user registration, sign-in, sessions, and OAuth.
- [**Study Sets and Folders**](./studysets_and_folders.md): Explains how study sets are created, organized, and shared.
- [**Terms and Images**](./terms_and_images.md): Details the structure of terms and how images are managed via GraphQL and REST.
- [**Progress and History**](./progress_and_history.md): Describes how user mastery and review history are tracked.
- [**Practice Tests**](./practice_tests.md): Outlines how practice tests are recorded and structured.

## Core Technologies

- **Backend:** Go (Golang)
- **Database:** PostgreSQL
- **GraphQL:** GQLGen
- **REST:** Chi (router)
- **Storage:** AWS S3 (for images)
- **Authentication:** Custom session management and Google OAuth
