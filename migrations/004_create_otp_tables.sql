-- Migration: Update OTP Tables for THSMS Integration
-- Description: Add columns to existing OTP tables for THSMS support

-- Add columns to otp_codes if they don't exist
ALTER TABLE public.otp_codes
ADD COLUMN IF NOT EXISTS blocked_until TIMESTAMP WITH TIME ZONE;

-- Add columns to otp_audit_log if they don't exist
ALTER TABLE public.otp_audit_log
ADD COLUMN IF NOT EXISTS action VARCHAR(50),
ADD COLUMN IF NOT EXISTS status VARCHAR(50),
ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES public.users(id) ON DELETE CASCADE;

-- Create index for audit log if not exists
CREATE INDEX IF NOT EXISTS idx_audit_action ON public.otp_audit_log(action);
CREATE INDEX IF NOT EXISTS idx_audit_status ON public.otp_audit_log(status);
CREATE INDEX IF NOT EXISTS idx_audit_user_id ON public.otp_audit_log(user_id);
