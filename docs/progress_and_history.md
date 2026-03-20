# Progress and History

Progress tracking allows users to see their mastery of terms and history of learning sessions.

## Core Entities

### Term Progress
User progress for each individual term is tracked in the `public.term_progress` table.
- **ID:** A unique UUID.
- **Term ID:** Links the progress record to a specific term.
- **User ID:** Links the progress to the user who studied the term.
- **Review Stats:** Tracks first and last review dates and the total number of reviews for both the "term" and "definition" sides.
- **Correct/Incorrect Counts:** Stores the cumulative number of correct and incorrect answers for each side.
- **Leitner System Boxes:** Tracks the mastery level (1-5) using the Leitner system.

### Term Progress History
The history of progress is tracked in the `public.term_progress_history` table.
- **ID:** A unique UUID.
- **Timestamp:** The date and time of the progress update.
- **Term ID:** Links the history record to a specific term.
- **User ID:** Links the history to the user.
- **Correct/Incorrect Counts:** Stores the counts at that specific point in time.

### Confusion Pairs
When a user incorrectly answers a question by selecting the wrong term, it's recorded in the `public.term_confusion_pairs` table.
- **ID:** A unique UUID.
- **User ID:** Links the confusion pair to the user.
- **Term ID:** The original term that was being tested.
- **Confused Term ID:** The term that was incorrectly selected.
- **Answered With:** Indicates which side was used for the question (`TERM` or `DEF`).
- **Confused Count:** The number of times this specific error occurred.
- **Last Confused At:** A timestamp of the most recent occurrence.

## Key Operations (GraphQL)

### Updating Progress
User progress is typically updated after a learning session:
- **Mutation:** `updateTermProgress(termProgress: [TermProgressInput!]!): [TermProgress!]`

The input includes:
- `termId`: The term whose progress is being updated.
- `termReviewedAt`, `defReviewedAt`: The timestamp of the review.
- `termLeitnerSystemBox`, `defLeitnerSystemBox`: The new Leitner box value.
- `termCorrectIncrease`, `termIncorrectIncrease`: The amount to increment the correct/incorrect counts.

This mutation updates the `public.term_progress` record and creates a new entry in `public.term_progress_history`.

### Recording Confusions
Incorrect answers can be recorded to help users identify their common mistakes:
- **Mutation:** `recordConfusedTerms(confusedTerms: [TermConfusionPairInput]): Boolean`

The input specifies which terms were confused and the type of question. This mutation updates or creates a record in `public.term_confusion_pairs`.

## Retrieving Progress
Progress information is integrated into the study set and term GraphQL queries, allowing the UI to display mastery levels and history alongside the terms themselves.
- `Term.progress`: Returns the current progress for the term.
- `Term.progressHistory`: Returns the history of progress for the term.
- `Term.topConfusionPairs`: Returns the most frequent confusion pairs where the current term was the correct answer.
- `Term.topReverseConfusionPairs`: Returns the most frequent confusion pairs where the current term was incorrectly chosen.
