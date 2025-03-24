-- Update admin password to '123456'
UPDATE users 
SET password = '$2a$10$iBJJ1cMetVj5uZoOQa4n/OgI8JvRnpGisyVkZoRPGrMRnK40m6Hi2' 
WHERE username = 'admin'; 