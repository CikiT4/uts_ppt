-- ============================================================
-- Legal Consultation Platform - Database Migration
-- Version: 1.0.0
-- ============================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- for full-text search

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE user_role AS ENUM ('client', 'lawyer', 'admin');
CREATE TYPE consultation_status AS ENUM ('pending', 'confirmed', 'ongoing', 'completed', 'cancelled');
CREATE TYPE payment_status AS ENUM ('pending', 'uploaded', 'verified', 'rejected', 'refunded');
CREATE TYPE payment_method AS ENUM ('bank_transfer', 'e_wallet', 'credit_card');
CREATE TYPE schedule_status AS ENUM ('available', 'booked', 'unavailable');
CREATE TYPE document_type AS ENUM ('consultation_doc', 'payment_proof', 'case_file', 'profile_photo', 'license');
CREATE TYPE message_type AS ENUM ('text', 'file', 'image');

-- ============================================================
-- TABLE: users (Base user table for all roles)
-- ============================================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role user_role NOT NULL DEFAULT 'client',
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    profile_photo_url VARCHAR(500),
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

-- ============================================================
-- TABLE: lawyers (Extended profile for lawyer role)
-- ============================================================
CREATE TABLE lawyers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    license_number VARCHAR(100) UNIQUE NOT NULL,
    specialization VARCHAR(255)[] NOT NULL DEFAULT '{}',
    years_of_experience INTEGER NOT NULL DEFAULT 0,
    education TEXT,
    bio TEXT,
    office_address TEXT,
    city VARCHAR(100),
    province VARCHAR(100),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    consultation_fee_per_hour DECIMAL(15, 2) NOT NULL DEFAULT 0,
    is_available BOOLEAN NOT NULL DEFAULT true,
    rating DECIMAL(3, 2) NOT NULL DEFAULT 0.00,
    total_reviews INTEGER NOT NULL DEFAULT 0,
    total_consultations INTEGER NOT NULL DEFAULT 0,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    verification_document_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lawyers_user_id ON lawyers(user_id);
CREATE INDEX idx_lawyers_city ON lawyers(city);
CREATE INDEX idx_lawyers_rating ON lawyers(rating DESC);
CREATE INDEX idx_lawyers_fee ON lawyers(consultation_fee_per_hour);
CREATE INDEX idx_lawyers_specialization ON lawyers USING GIN(specialization);

-- ============================================================
-- TABLE: clients (Extended profile for client role)
-- ============================================================
CREATE TABLE clients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    date_of_birth DATE,
    address TEXT,
    city VARCHAR(100),
    province VARCHAR(100),
    occupation VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_clients_user_id ON clients(user_id);

-- ============================================================
-- TABLE: lawyer_schedules (Available time slots)
-- ============================================================
CREATE TABLE lawyer_schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lawyer_id UUID NOT NULL REFERENCES lawyers(id) ON DELETE CASCADE,
    day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6), -- 0=Sunday
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_schedule_time CHECK (end_time > start_time)
);

CREATE INDEX idx_lawyer_schedules_lawyer_id ON lawyer_schedules(lawyer_id);

-- ============================================================
-- TABLE: consultations (Booking & consultation sessions)
-- ============================================================
CREATE TABLE consultations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE RESTRICT,
    lawyer_id UUID NOT NULL REFERENCES lawyers(id) ON DELETE RESTRICT,
    schedule_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    duration_hours DECIMAL(4, 2) NOT NULL,
    status consultation_status NOT NULL DEFAULT 'pending',
    case_description TEXT NOT NULL,
    case_type VARCHAR(100),
    consultation_fee DECIMAL(15, 2) NOT NULL,
    platform VARCHAR(50) DEFAULT 'chat', -- chat, video_call, in_person
    meeting_link VARCHAR(500),
    notes TEXT,
    cancelled_reason TEXT,
    cancelled_by UUID REFERENCES users(id),
    confirmed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_consultations_client_id ON consultations(client_id);
CREATE INDEX idx_consultations_lawyer_id ON consultations(lawyer_id);
CREATE INDEX idx_consultations_status ON consultations(status);
CREATE INDEX idx_consultations_schedule_date ON consultations(schedule_date);

-- ============================================================
-- TABLE: payments (Payment records)
-- ============================================================
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    consultation_id UUID NOT NULL UNIQUE REFERENCES consultations(id) ON DELETE RESTRICT,
    amount DECIMAL(15, 2) NOT NULL,
    payment_method payment_method,
    payment_status payment_status NOT NULL DEFAULT 'pending',
    bank_name VARCHAR(100),
    account_number VARCHAR(50),
    transfer_reference VARCHAR(100),
    payment_proof_url VARCHAR(500),
    payment_date TIMESTAMPTZ,
    verified_by UUID REFERENCES users(id),
    verified_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_consultation_id ON payments(consultation_id);
CREATE INDEX idx_payments_status ON payments(payment_status);

-- ============================================================
-- TABLE: reviews (Lawyer reviews by clients)
-- ============================================================
CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    consultation_id UUID NOT NULL UNIQUE REFERENCES consultations(id) ON DELETE RESTRICT,
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE RESTRICT,
    lawyer_id UUID NOT NULL REFERENCES lawyers(id) ON DELETE RESTRICT,
    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    is_anonymous BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reviews_lawyer_id ON reviews(lawyer_id);
CREATE INDEX idx_reviews_client_id ON reviews(client_id);

-- ============================================================
-- TABLE: documents (File management)
-- ============================================================
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    uploader_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    consultation_id UUID REFERENCES consultations(id) ON DELETE CASCADE,
    document_type document_type NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    file_size INTEGER NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_documents_uploader_id ON documents(uploader_id);
CREATE INDEX idx_documents_consultation_id ON documents(consultation_id);
CREATE INDEX idx_documents_type ON documents(document_type);

-- ============================================================
-- TABLE: chat_messages (Real-time messaging)
-- ============================================================
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    consultation_id UUID NOT NULL REFERENCES consultations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    message_type message_type NOT NULL DEFAULT 'text',
    content TEXT,
    file_url VARCHAR(500),
    file_name VARCHAR(255),
    is_read BOOLEAN NOT NULL DEFAULT false,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chat_messages_consultation_id ON chat_messages(consultation_id);
CREATE INDEX idx_chat_messages_sender_id ON chat_messages(sender_id);
CREATE INDEX idx_chat_messages_created_at ON chat_messages(created_at);

-- ============================================================
-- TABLE: notifications
-- ============================================================
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    reference_id UUID,
    reference_type VARCHAR(50),
    is_read BOOLEAN NOT NULL DEFAULT false,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);

-- ============================================================
-- FUNCTION: Update updated_at timestamp
-- ============================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- ============================================================
-- TRIGGERS: Auto-update timestamps
-- ============================================================
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_lawyers_updated_at BEFORE UPDATE ON lawyers FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_clients_updated_at BEFORE UPDATE ON clients FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_lawyer_schedules_updated_at BEFORE UPDATE ON lawyer_schedules FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_consultations_updated_at BEFORE UPDATE ON consultations FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_reviews_updated_at BEFORE UPDATE ON reviews FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- ============================================================
-- FUNCTION: Update lawyer rating after review
-- ============================================================
CREATE OR REPLACE FUNCTION update_lawyer_rating()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE lawyers
    SET 
        rating = (
            SELECT ROUND(AVG(rating)::numeric, 2)
            FROM reviews
            WHERE lawyer_id = NEW.lawyer_id
        ),
        total_reviews = (
            SELECT COUNT(*)
            FROM reviews
            WHERE lawyer_id = NEW.lawyer_id
        )
    WHERE id = NEW.lawyer_id;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_update_lawyer_rating
AFTER INSERT OR UPDATE ON reviews
FOR EACH ROW EXECUTE PROCEDURE update_lawyer_rating();

-- ============================================================
-- SEED DATA: Admin user
-- ============================================================
INSERT INTO users (email, password_hash, role, full_name, is_verified, is_active)
VALUES (
    'admin@legalconsult.id',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj4oS8PXkMO2', -- password: Admin@123
    'admin',
    'System Administrator',
    true,
    true
);
