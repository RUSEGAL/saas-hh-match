# GitHub Actions Environments

## Required Secrets

### For CI/CD Pipeline

| Secret | Description | Where to get |
|--------|-------------|--------------|
| `BOT_TOKEN` | Telegram bot token | @BotFather |
| `AI_API_KEY` | DeepSeek API key | platform.deepseek.com |
| `DB_PASSWORD` | PostgreSQL password | Generate secure string |
| `JWT_SECRET` | JWT signing secret | Generate 64-char random string |
| `REDIS_PASSWORD` | Redis password (optional) | Generate secure string |
| `WEBHOOK_SECRET` | Telegram webhook secret | Generate random string |

### For Production AWS Deployment (optional)

| Secret | Description |
|--------|-------------|
| `AWS_ACCESS_KEY_ID` | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key |
| `AWS_REGION` | AWS region (e.g., eu-central-1) |

### For Slack Notifications (optional)

| Secret | Description |
|--------|-------------|
| `SLACK_WEBHOOK_URL` | Slack incoming webhook URL |

## Setting up Secrets

1. Go to Settings > Secrets and variables > Actions
2. Add new repository secret for each value
3. For production, also add to Environment secrets

## GitHub Container Registry

Images are pushed to `ghcr.io/{owner}/{repo}/`

Example:
```
ghcr.io/myusername/resume-bot-go-api:latest
```

## Environment Protection

Production environment requires:
- [ ] Approval from required reviewers
- [ ] No concurrent deployments
