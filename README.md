# 🐾 Pet Theory – Serverless REST API with Go & Google Cloud Run

This project is part of the Google Cloud SkillBoosts (Coursera) Lab – GSP761: Developing a REST API with Go and Cloud Run.

It shows how to build and deploy a secure REST API using Go, Firestore, and Cloud Run.

---

## 🌟 Overview
Pet Theory is a fictional veterinary clinic chain.  
This API allows insurance companies to view customer treatment costs without exposing personal data.

### 🎯 Objectives
- Build REST API in Go  
- Connect to Firestore Database  
- Deploy using Cloud Run  
- Handle data securely with CORS and JSON

---

## ⚙️ Commands used
```bash
GOOS=linux GOARCH=amd64 go build -o server
gcloud builds submit --tag gcr.io/$GOOGLE_CLOUD_PROJECT/rest-api:0.1



