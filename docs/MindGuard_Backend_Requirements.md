# MediaDetox Backend Requirements

## Product Vision

**MediaDetox** 

A backend platform that detects compulsive application usage patterns and generates alerts, metrics, and recommendations focused on digital well-being.

---

# MVP (Version 1)

## Goals

The system must:

- Receive application usage events.
- Store usage events.
- Analyze behavioral patterns.
- Generate alerts.
- Expose operational metrics for Prometheus.
- Publish notification events to a Notification Center.

**Note:** No blocking functionality will be implemented in the MVP.

---

# Functional Requirements

## FR-001 Register Events

The system must allow clients to register activity events.

Example:

```json
{
  "user_id": "123",
  "app_name": "instagram",
  "event_type": "open",
  "timestamp": "2026-06-23T10:00:00Z"
}
```

Supported event types:

- open
- close

---

## FR-002 Query Events

Retrieve historical events.

```http
GET /users/{id}/events
```

Supported filters:

- start_date
- end_date
- app_name

---

## FR-003 Generate Alerts

The system must generate alerts when configured behavioral patterns are detected.

Examples:

- excessive opens
- night usage
- checking loops

Generated alerts must be persisted and published to the Notification Center so downstream channels can send push notifications, email, or other user-facing messages.

---

## FR-004 Query Alerts

```http
GET /users/{id}/alerts
```

Response:

```json
[
  {
    "type": "checking_loop",
    "severity": "medium"
  }
]
```

---

## FR-005 Query Usage Summary

```http
GET /users/{id}/metrics
```

Example:

```json
{
  "opens_today": 54,
  "most_used_app": "instagram",
  "alerts_today": 3
}
```

This endpoint returns product/domain usage summaries for a user. Operational service metrics must be exposed separately for Prometheus scraping.

---

## FR-010 Expose Prometheus Metrics

```http
GET /metrics
```

The endpoint must expose Prometheus-compatible operational metrics such as request counts, request duration, processed events, generated alerts, and repository errors.

---

## FR-006 Configure Rules

Create custom detection rules.

```http
POST /rules
```

Example:

```json
{
  "app_name": "instagram",
  "max_opens": 15,
  "window_minutes": 60
}
```

---

# Rule Engine Requirements

## FR-007 Checking Loop Detection

Detect repetitive application openings within a short time window.

Example:

```text
Instagram

10:00
10:02
10:04
10:05
10:07
10:08
```

Result:

```text
Checking Loop Alert
```

---

## FR-008 Excessive Usage Detection

Detect excessive daily usage.

Example:

```text
Instagram opened 85 times today
```

---

## FR-009 Night Usage Detection

Detect activity during configured sleeping hours.

Example:

```text
03:12
03:40
04:01
```

---

# Data Model

## User

```text
id
name
created_at
```

---

## Event

```text
id
user_id
app_name
event_type
timestamp
```

---

## Alert

```text
id
user_id
alert_type
severity
message
created_at
```

---

## Rule

```text
id
name
app_name
max_opens
window_minutes
enabled
```

---

# Non-Functional Requirements

## NFR-001 REST API

All operations must be exposed through a REST API.

---

## NFR-002 PostgreSQL

PostgreSQL will be the primary persistence layer.

---

## NFR-003 Docker

The application must be runnable through Docker Compose.

---

## NFR-004 Swagger / OpenAPI

API documentation must be available through OpenAPI.

---

## NFR-005 Logging

Use structured logging.

Example:

```json
{
  "level": "info",
  "event": "alert_generated"
}
```

---

## NFR-006 Prometheus Metrics

Prometheus must collect service metrics by scraping the backend `/metrics` endpoint. Metrics must not be pushed through SQS or persisted as application records.

---

## NFR-007 Notification Center

The backend must not call mobile push providers directly from the rule engine. Alert generation must publish a notification event to a Notification Center, which owns push delivery through providers such as FCM or APNs.

---

## NFR-008 Concurrency

Events must be processed asynchronously.

Flow:

```text
API
 ↓
Store Event
 ↓
Channel
 ↓
Worker
 ↓
Rule Engine
 ↓
Alerts
 ↓
Notification Center
```

---

# MVP API Endpoints

```http
POST   /events

GET    /users/{id}/events

GET    /users/{id}/alerts

GET    /users/{id}/metrics

GET    /metrics

POST   /rules

GET    /rules

PUT    /rules/{id}

DELETE /rules/{id}
```

---

# Future Roadmap (Version 2)

The following features are intentionally out of scope for the MVP:

- JWT Authentication
- WebSocket notifications
- Recommendation engine
- Behavioral Risk Score (0-100)
- Machine Learning pattern detection
- Android integration
- iOS integration
- Automatic application blocking

---

# Recommended Development Order

1. Event ingestion API.
2. PostgreSQL persistence.
3. Rule Engine.
4. Alert generation.
5. Metrics aggregation.
6. Asynchronous processing with goroutines and channels.
7. Swagger documentation.
8. Docker Compose setup.

