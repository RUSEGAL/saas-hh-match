#!/bin/sh

echo "Waiting for NATS to be ready..."
until curl -s http://nats:8222/healthz > /dev/null 2>&1; do
  sleep 1
done
sleep 2

echo "Creating RESUME_ANALYZE stream..."
curl -s -X PUT http://nats:8222/jsz/streams -H "Content-Type: application/json" -d '{
  "name": "RESUME_ANALYZE",
  "subjects": ["resume.analyze"],
  "retention": "workqueue",
  "storage": "file"
}' > /dev/null 2>&1 || echo "Stream may already exist"

echo "Creating VACANCY_MATCH stream..."
curl -s -X PUT http://nats:8222/jsz/streams -H "Content-Type: application/json" -d '{
  "name": "VACANCY_MATCH",
  "subjects": ["vacancy.match"],
  "retention": "workqueue",
  "storage": "file"
}' > /dev/null 2>&1 || echo "Stream may already exist"

echo "NATS streams configured successfully"
exit 0
