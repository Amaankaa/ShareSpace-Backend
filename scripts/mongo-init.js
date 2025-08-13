// MongoDB initialization script for ShareSpace
print('Starting ShareSpace database initialization...');

// Switch to ShareSpace database
db = db.getSiblingDB('sharespace');

// Create application user
db.createUser({
  user: 'sharespace_app',
  pwd: 'sharespace_password', // Change this in production
  roles: [
    {
      role: 'readWrite',
      db: 'sharespace'
    }
  ]
});

// Create collections with validation
print('Creating users collection...');
db.createCollection('users', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['username', 'email', 'password', 'fullname', 'role', 'isVerified'],
      properties: {
        username: {
          bsonType: 'string',
          minLength: 3,
          maxLength: 30,
          description: 'Username must be a string between 3-30 characters'
        },
        email: {
          bsonType: 'string',
          pattern: '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$',
          description: 'Email must be a valid email address'
        },
        password: {
          bsonType: 'string',
          minLength: 8,
          description: 'Password must be at least 8 characters'
        },
        fullname: {
          bsonType: 'string',
          minLength: 2,
          maxLength: 100,
          description: 'Full name must be between 2-100 characters'
        },
        role: {
          bsonType: 'string',
          enum: ['admin', 'user'],
          description: 'Role must be either admin or user'
        },
        isVerified: {
          bsonType: 'bool',
          description: 'isVerified must be a boolean'
        },
        displayName: {
          bsonType: 'string',
          maxLength: 50,
          description: 'Display name must be max 50 characters'
        },
        isAnonymous: {
          bsonType: 'bool',
          description: 'isAnonymous must be a boolean'
        },
        isMentor: {
          bsonType: 'bool',
          description: 'isMentor must be a boolean'
        },
        isMentee: {
          bsonType: 'bool',
          description: 'isMentee must be a boolean'
        },
        mentorshipTopics: {
          bsonType: 'array',
          items: {
            bsonType: 'string'
          },
          description: 'mentorshipTopics must be an array of strings'
        },
        availableForMentoring: {
          bsonType: 'bool',
          description: 'availableForMentoring must be a boolean'
        }
      }
    }
  }
});

print('Creating mentorship_requests collection...');
db.createCollection('mentorship_requests', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['menteeId', 'mentorId', 'status', 'topics', 'createdAt', 'updatedAt'],
      properties: {
        menteeId: {
          bsonType: 'objectId',
          description: 'menteeId must be a valid ObjectId'
        },
        mentorId: {
          bsonType: 'objectId',
          description: 'mentorId must be a valid ObjectId'
        },
        status: {
          bsonType: 'string',
          enum: ['pending', 'accepted', 'rejected', 'canceled'],
          description: 'status must be one of: pending, accepted, rejected, canceled'
        },
        topics: {
          bsonType: 'array',
          minItems: 1,
          items: {
            bsonType: 'string'
          },
          description: 'topics must be a non-empty array of strings'
        },
        message: {
          bsonType: 'string',
          maxLength: 500,
          description: 'message must be max 500 characters'
        }
      }
    }
  }
});

print('Creating mentorship_connections collection...');
db.createCollection('mentorship_connections', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['menteeId', 'mentorId', 'requestId', 'status', 'topics', 'startedAt', 'createdAt', 'updatedAt'],
      properties: {
        menteeId: {
          bsonType: 'objectId',
          description: 'menteeId must be a valid ObjectId'
        },
        mentorId: {
          bsonType: 'objectId',
          description: 'mentorId must be a valid ObjectId'
        },
        requestId: {
          bsonType: 'objectId',
          description: 'requestId must be a valid ObjectId'
        },
        status: {
          bsonType: 'string',
          enum: ['active', 'paused', 'completed', 'ended'],
          description: 'status must be one of: active, paused, completed, ended'
        },
        topics: {
          bsonType: 'array',
          minItems: 1,
          items: {
            bsonType: 'string'
          },
          description: 'topics must be a non-empty array of strings'
        },
        menteeRating: {
          bsonType: 'int',
          minimum: 1,
          maximum: 5,
          description: 'menteeRating must be between 1-5'
        },
        mentorRating: {
          bsonType: 'int',
          minimum: 1,
          maximum: 5,
          description: 'mentorRating must be between 1-5'
        }
      }
    }
  }
});

// Create indexes for performance
print('Creating indexes...');

// Users collection indexes
db.users.createIndex({ 'username': 1 }, { unique: true });
db.users.createIndex({ 'email': 1 }, { unique: true });
db.users.createIndex({ 'displayName': 1 }, { unique: true, sparse: true });
db.users.createIndex({ 'isMentor': 1, 'availableForMentoring': 1 });
db.users.createIndex({ 'isMentee': 1 });
db.users.createIndex({ 'mentorshipTopics': 1 });

// Mentorship requests indexes
db.mentorship_requests.createIndex({ 'menteeId': 1 });
db.mentorship_requests.createIndex({ 'mentorId': 1 });
db.mentorship_requests.createIndex({ 'status': 1 });
db.mentorship_requests.createIndex({ 'menteeId': 1, 'mentorId': 1, 'status': 1 });
db.mentorship_requests.createIndex({ 'createdAt': 1 });
db.mentorship_requests.createIndex({ 'topics': 1 });

// Mentorship connections indexes
db.mentorship_connections.createIndex({ 'menteeId': 1 });
db.mentorship_connections.createIndex({ 'mentorId': 1 });
db.mentorship_connections.createIndex({ 'requestId': 1 }, { unique: true });
db.mentorship_connections.createIndex({ 'status': 1 });
db.mentorship_connections.createIndex({ 'startedAt': 1 });
db.mentorship_connections.createIndex({ 'topics': 1 });

// Compound indexes for common queries
db.mentorship_connections.createIndex({ 'menteeId': 1, 'status': 1 });
db.mentorship_connections.createIndex({ 'mentorId': 1, 'status': 1 });

print('ShareSpace database initialization completed successfully!');
print('Collections created: users, mentorship_requests, mentorship_connections');
print('Indexes created for optimal query performance');
print('Application user created: sharespace_app');
