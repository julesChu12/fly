-- Create databases for the three new services
CREATE DATABASE IF NOT EXISTS hermes;
CREATE DATABASE IF NOT EXISTS kratos;
CREATE DATABASE IF NOT EXISTS plutus;

-- Grant permissions
GRANT ALL PRIVILEGES ON hermes.* TO 'custos'@'%';
GRANT ALL PRIVILEGES ON kratos.* TO 'custos'@'%';
GRANT ALL PRIVILEGES ON plutus.* TO 'custos'@'%';

FLUSH PRIVILEGES;