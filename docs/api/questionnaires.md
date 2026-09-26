# Questionnaire API

- `GET|POST /api/v1/applications/{applicationID}/questionnaires`
- `GET|PATCH|DELETE /api/v1/applications/{applicationID}/questionnaires/{questionnaireID}`
- `GET|POST /api/v1/questionnaires/{questionnaireID}/questions`
- `PATCH|DELETE /api/v1/questionnaires/{questionnaireID}/questions/{questionID}`
- `GET|POST /api/v1/questions/{questionID}/answers`
- `PATCH /api/v1/questions/{questionID}/answers/{answerID}`
- `POST .../{answerID}/approve|reject`
- `POST /api/v1/questions/{questionID}/regenerate`

Suggested answers are not equivalent to approved answers.
