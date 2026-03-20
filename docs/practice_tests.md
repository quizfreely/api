# Practice Tests

Practice tests allow users to assess their knowledge of a study set by answering a series of questions.

## Core Entities

### Practice Tests
Information about each practice test is stored in the `public.practice_tests` table.
- **ID:** A unique UUID.
- **Timestamp:** The date and time the test was completed.
- **User ID:** Links the test to the user who took it.
- **Studyset ID:** Links the test to the study set.
- **Questions Correct:** The number of questions answered correctly.
- **Questions Total:** The total number of questions in the test.
- **Questions:** A JSONB field that stores the individual questions, answers, and correctness results.

## Key Operations (GraphQL)

### Recording a Test
After completing a practice test, the user's performance is recorded:
- **Mutation:** `recordPracticeTest(input: PracticeTestInput): PracticeTest`

The input includes:
- `studysetId`: The study set the test was based on.
- `questionsCorrect`: The total number of correct answers.
- `questionsTotal`: The total number of questions.
- `questions`: A detailed list of individual questions.

### Updating a Test
If needed, a practice test record can be updated (e.g., if a user manually marks a question as correct):
- **Mutation:** `updatePracticeTest(input: PracticeTestInput): PracticeTest`

### Retrieving Tests
A list of practice tests taken for a study set can be retrieved:
- `Studyset.practiceTests`: Returns a list of practice test results for the study set.

## Question Types
The `questions` field in the `PracticeTestInput` can contain various question types, each with its own structure:
- **MCQ (Multiple Choice):** Includes the term, the chosen answer, distractors, and the correct choice index.
- **True/False:** Includes the term, the user's choice, and whether it was correct.
- **Match:** Includes the term, the matched term, and the group identifier.
- **FRQ (Free Response):** Includes the term, the user's typed response, and whether the user manually marked it as correct.
