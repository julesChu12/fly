-- Create all databases required by Fly microservices
CREATE DATABASE IF NOT EXISTS custos DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS hermes DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS kratos DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS plutus DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS staff DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS appointments DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS items_dev DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS clotho_db DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create users and grant privileges
-- Custos service
CREATE USER IF NOT EXISTS 'custos'@'%' IDENTIFIED BY 'custospassword';
GRANT ALL PRIVILEGES ON custos.* TO 'custos'@'%';

-- Hermes service
CREATE USER IF NOT EXISTS 'hermes'@'%' IDENTIFIED BY 'hermespassword';
GRANT ALL PRIVILEGES ON hermes.* TO 'hermes'@'%';

-- Kratos service (使用 root 用户已存在，这里创建专用用户作为备选)
CREATE USER IF NOT EXISTS 'kratos'@'%' IDENTIFIED BY 'kratospassword';
GRANT ALL PRIVILEGES ON kratos.* TO 'kratos'@'%';

-- Plutus service
CREATE USER IF NOT EXISTS 'plutus'@'%' IDENTIFIED BY 'plutuspassword';
GRANT ALL PRIVILEGES ON plutus.* TO 'plutus'@'%';

-- Staff service (使用 root 用户已存在，这里创建专用用户作为备选)
CREATE USER IF NOT EXISTS 'staff'@'%' IDENTIFIED BY 'staffpassword';
GRANT ALL PRIVILEGES ON staff.* TO 'staff'@'%';

-- Appointments service
CREATE USER IF NOT EXISTS 'appointments'@'%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON appointments.* TO 'appointments'@'%';

-- Items service
CREATE USER IF NOT EXISTS 'fly_user'@'%' IDENTIFIED BY 'rootpassword';
GRANT ALL PRIVILEGES ON items_dev.* TO 'fly_user'@'%';

-- Clotho service
CREATE USER IF NOT EXISTS 'clotho'@'%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON clotho_db.* TO 'clotho'@'%';

-- Grant root user access to all databases (for development convenience)
GRANT ALL PRIVILEGES ON *.* TO 'root'@'%' WITH GRANT OPTION;

-- Apply changes
FLUSH PRIVILEGES;
