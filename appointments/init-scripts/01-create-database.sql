-- Create appointments database if it doesn't exist
CREATE DATABASE IF NOT EXISTS appointments DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Use the appointments database
USE appointments;

-- Create appointments table if it doesn't exist
CREATE TABLE IF NOT EXISTS appointments (
    id CHAR(36) PRIMARY KEY UNIQUE,
    customer_id CHAR(36) NOT NULL,
    staff_id CHAR(36) NOT NULL,
    service_id CHAR(36) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    notes TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reminder BOOLEAN DEFAULT TRUE,
    reminder_time TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,

    INDEX idx_customer_id (customer_id),
    INDEX idx_staff_id (staff_id),
    INDEX idx_service_id (service_id),
    INDEX idx_start_time (start_time),
    INDEX idx_status (status),
    INDEX idx_reminder_time (reminder_time),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at),
    INDEX idx_unique_id (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert some sample data for development
INSERT IGNORE INTO appointments (
    id, customer_id, staff_id, service_id, start_time, end_time, notes, status
) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440101', '550e8400-e29b-41d4-a716-446655440201', '550e8400-e29b-41d4-a716-446655440301', NOW() + INTERVAL 1 DAY, NOW() + INTERVAL 1 DAY + INTERVAL 1 HOUR, 'Regular checkup', 'confirmed'),
    ('550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440102', '550e8400-e29b-41d4-a716-446655440202', '550e8400-e29b-41d4-a716-446655440302', NOW() + INTERVAL 2 DAYS, NOW() + INTERVAL 2 DAYS + INTERVAL 30 MINUTES, 'Follow-up consultation', 'pending'),
    ('550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440103', '550e8400-e29b-41d4-a716-446655440203', '550e8400-e29b-41d4-a716-446655440303', NOW() + INTERVAL 3 DAYS, NOW() + INTERVAL 3 DAYS + INTERVAL 2 HOURS, 'Specialist consultation', 'confirmed');