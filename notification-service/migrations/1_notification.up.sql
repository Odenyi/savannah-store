CREATE TABLE notification (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,              -- who this notification is for
    type ENUM('sms', 'email') NOT NULL,   -- type of notification
    recipient VARCHAR(255) NOT NULL,      -- phone number or email address
    subject VARCHAR(255),                  -- only for email
    message TEXT NOT NULL,                 -- message content
    status ENUM('pending', 'sent', 'failed') DEFAULT 'pending', -- delivery status
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
