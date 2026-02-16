-- ============================================================================
-- ADD USERNAME AND RESOURCE_TYPE TO AUDIT_LOGS
-- ============================================================================
ALTER TABLE audit_logs
ADD COLUMN username VARCHAR(255),
ADD COLUMN resource_type VARCHAR(100);

-- Create index for faster queries
CREATE INDEX idx_audit_logs_username ON audit_logs(username);
CREATE INDEX idx_audit_logs_resource_type ON audit_logs(resource_type);

-- Update existing records to populate username from users table
UPDATE audit_logs al
SET username = u.username
FROM users u
WHERE al.actor_id = u.id AND al.username IS NULL;
