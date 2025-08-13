# ShareSpace

ShareSpace is a web platform designed to connect junior university students with senior students and mentors, helping them navigate academic, financial, and relationship challenges. It also provides curated resources and a supportive community for mental health and personal growth.

---

## 🚀 Features

### User Management

* User registration & login (JWT-based authentication)
* Forgot password & OTP verification
* Role-based access (User, Admin, Mentor)
* Promote/Demote user roles (admin only)

### Mentorship

* Send & manage mentorship requests
* Accept/Reject connections
* View active mentorship relationships

### Discussion Board

* Create, view, edit, and delete posts
* Comment & like functionality
* Category-based filtering

### Resource Sharing

* Share documents, links, and guides
* Search & browse resources

### Messaging

* One-to-one real-time chat between mentor & mentee
* Message history storage

### Mental Health Support

* Curated articles
* Option to book counseling/mentorship sessions

---

## 🛠️ Tech Stack

**Frontend**

* React (TypeScript)
* TailwindCSS
* React Query (API data fetching)
* Axios (HTTP client)

**Backend**

* Go (Golang)
* MongoDB
* Gin/Fiber (HTTP framework)
* JWT (Authentication)
* Clean Architecture

**Other Tools**

* Docker (Containerization)
* GitHub Actions (CI/CD)
* WebSockets (Real-time messaging)

---

## 📚 Project Structure (Backend - Clean Architecture)

```
ShareSpace/
│── Delivery/           # HTTP handlers/controllers
│── Domain/             # Entities & interfaces
│── Infrastructure/     # MongoDB, email service, etc.
│── Repositories/       # DB repository implementations
│── Usecases/           # Business logic
│── main.go              # App entry point
```

---

## 💾 Installation & Setup

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/sharespace.git
cd sharespace
```

### 2. Backend Setup

```bash
cd backend
cp .env.example .env   # Add your MongoDB URI, JWT secret, etc.
go mod tidy
go run main.go
```

### 3. Frontend Setup

```bash
cd frontend
cp .env.example .env   # Add your API base URL, WebSocket URL, etc.
npm install
npm run dev
```

---

## 🤚 Running Tests

**Backend**

```bash
cd backend
go test ./...
```

**Frontend**

```bash
cd frontend
npm run test
```

---

## 🚀 Deployment

Using Docker:

```bash
docker-compose up --build
```

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch:

   ```bash
   git checkout -b feature-name
   ```
3. Commit changes and push:

   ```bash
   git commit -m "Add feature"
   git push origin feature-name
   ```
4. Open a pull request

---

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.