# AI Usage Disclosure

- Tool used: ChatGPT.
- Used for: understanding requirements, planning implementation, drafting Go code and documentation, and identifying concurrency and idempotency risks.
- Suggestion changed: deployment to Vercel was not prioritized because the assignment explicitly says deployment is not required and in-memory state is unsuitable for serverless persistence.
- One thing AI helped explain: why event deduplication must use a global event ID and why unique opens must be scoped by campaign and contact.
- AI-generated code must be reviewed and tested locally. Any compile or test failures will be corrected before submission.
