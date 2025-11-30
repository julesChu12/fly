-- Create all databases required by Fly microservices
CREATE DATABASE IF NOT EXISTS custos DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS hermes DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS kratos DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS plutus DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS staff DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS appointments DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS items DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create users and grant privileges
CREATE USER IF NOT EXISTS 'custos'@'%' IDENTIFIED BY 'custospassword';
CREATE USER IF NOT EXISTS 'hermes'@'%' IDENTIFIED BY 'hermespassword';
CREATE USER IF NOT EXISTS 'kratos'@'%' IDENTIFIED BY 'kratospassword';
CREATE USER IF NOT EXISTS 'plutus'@'%' IDENTIFIED BY 'plutuspassword';
CREATE USER IF NOT EXISTS 'staff'@'%' IDENTIFIED BY 'staffpassword';
CREATE USER IF NOT EXISTS 'appointments'@'%' IDENTIFIED BY 'appointmentspassword';
CREATE USER IF NOT EXISTS 'items'@'%' IDENTIFIED BY 'itemspassword';

-- Grant privileges on respective databases
GRANT ALL PRIVILEGES ON custos.* TO 'custos'@'%';
GRANT ALL PRIVILEGES ON hermes.* TO 'hermes'@'%';
GRANT ALL PRIVILEGES ON kratos.* TO 'kratos'@'%';
GRANT ALL PRIVILEGES ON plutus.* TO 'plutus'@'%';
GRANT ALL PRIVILEGES ON staff.* TO 'staff'@'%';
GRANT ALL PRIVILEGES ON appointments.* TO 'appointments'@'%';
GRANT ALL PRIVILEGES ON items.* TO 'items'@'%';

-- Apply changes
FLUSH PRIVILEGES;