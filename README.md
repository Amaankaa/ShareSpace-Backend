# 📝 G6 Blog Starter Project

The **G6 Blog Starter** is a backend API built to power a feature-rich blogging platform. It supports user management, blog post operations, advanced search and filtering, AI content suggestions, and role-based access control — all built with performance and scalability in mind.

---

## 🚀 Features

- 🔐 **User Authentication & Authorization** (JWT-based, with refresh tokens)  
- 👤 **User Roles** (Admin & User, with role promotion/demotion)  
- 📝 **Blog CRUD** (Create, Read with pagination/search/filter, Update, Delete)  
- 🧠 **AI-Powered Blog Suggestions** based on keywords or topics  
- 📈 **Popularity Tracking** (views, likes/dislikes, comments)  
- 🔍 **Advanced Blog Filtering** (by tags, date, or popularity)  
- 📬 **Forgot Password + Reset via Email**  
- 🧾 **Profile Management** (bio, profile picture, contact info)

---

## ⚙️ API Design Goals

- RESTful and intuitive endpoints  
- Secure user auth with hashed passwords (`bcrypt`)  
- Token-based session handling with access & refresh token flows  
- Clean role-based permission handling (RBAC via middleware)  
- Full Postman API documentation with examples and error responses  
- Lightweight, fast, and concurrent using Go’s `goroutines` & `sync` patterns

---

## 🧠 AI Integration

Integrates a basic AI assistant that allows users to:  
- Get content suggestions for their blogs  
- Generate full blog outlines based on input topics or keywords

---

## 📦 Tech Stack

- **Language:** Go (Golang)  
- **Database:** MongoDB (or any pluggable DB)  
- **Auth:** JWT (Access & Refresh Tokens)  
- **AI Integration:** Placeholder AI endpoint for future LLM integrations  
- **Docs:** Postman collection with example requests/responses

---

## 📌 Non-Functional Highlights

- ⚡ **High Performance:** Optimized queries, paginated responses  
- 🔒 **Security:** `bcrypt` password hashing, JWT token validation  
- ⚖️ **Scalability:** Built with Go’s concurrency model  
- 🧰 **Maintainability:** Clean architecture with clear separation of concerns

---

## 📖 Project Status

This is a starter template meant to kickstart the development of scalable blog platforms. Ideal for learning advanced Go practices or rapid prototyping.

---

## 📄 API Docs

All endpoints are fully documented in Postman and include:  
- ✅ Request and response samples  
- ❌ Error handling scenarios  
- 🔑 Role-based access details
