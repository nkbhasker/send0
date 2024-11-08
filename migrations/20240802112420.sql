-- Create enum type "campaign_status"
CREATE TYPE "public"."campaign_status" AS ENUM ('DRAFT', 'SCHEDULED', 'RUNNING', 'COMPLETED', 'CANCELED');
-- Create enum type "domain_status"
CREATE TYPE "public"."domain_status" AS ENUM ('PENDING', 'ACTIVE', 'INACTIVE');
-- Create enum type "event_type"
CREATE TYPE "public"."event_type" AS ENUM ('EMAIL_SENT', 'EMAIL_DELIVERED', 'EMAIL_OPENED', 'EMAIL_CLICKED', 'EMAIL_BOUNCED', 'EMAIL_UNSUBSCRIBED', 'EMAIL_REPORTED', 'EMAIL_REJECTED');
-- Create enum type "identity_provider"
CREATE TYPE "public"."identity_provider" AS ENUM ('LOCAL', 'GOOGLE', 'GITHUB');
-- Create enum type "team_user_status"
CREATE TYPE "public"."team_user_status" AS ENUM ('ACTIVE', 'INACTIVE', 'PENDING');
-- Create "clients" table
CREATE TABLE "public"."clients" (
  "id" bigint NOT NULL,
  "updated_at" timestamptz NULL,
  "description" character varying(255) NULL,
  "secret" character varying(255) NOT NULL,
  "last_used_at" timestamptz NULL,
  "permissions" jsonb NOT NULL DEFAULT '[]',
  "workspace_id" bigint NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "domains" table
CREATE TABLE "public"."domains" (
  "id" bigint NOT NULL,
  "updated_at" timestamptz NULL,
  "name" text NOT NULL,
  "region" text NOT NULL,
  "status" "public"."domain_status" NOT NULL DEFAULT 'PENDING',
  "dkim_records" jsonb NOT NULL,
  "spf_records" jsonb NOT NULL,
  "dmarc_records" jsonb NOT NULL,
  "private_key" boolean NOT NULL,
  "organization_id" bigint NOT NULL,
  "workspace_id" bigint NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "emails" table
CREATE TABLE "public"."emails" (
  "id" bigint NOT NULL,
  "updated_at" timestamptz NULL,
  "message_id" uuid NULL,
  "from_address" text NULL,
  "recipients" jsonb NOT NULL,
  "cc_recipients" jsonb NOT NULL,
  "bcc_recipients" jsonb NOT NULL,
  "subject" text NULL,
  "content" text NULL,
  "status" text NULL,
  "delay" bigint NULL,
  "delay_time_zone" text NULL,
  "sent_at" timestamptz NULL,
  "attachments" jsonb NOT NULL,
  "organization_id" bigint NOT NULL,
  "workspace_id" bigint NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "events" table
CREATE TABLE "public"."events" (
  "id" bigint NOT NULL,
  "updated_at" timestamptz NULL,
  "event_type" "public"."event_type" NOT NULL,
  "receipients" jsonb NOT NULL,
  "meta_data" jsonb NOT NULL,
  "organization_id" bigint NOT NULL,
  "workspace_id" bigint NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "organizations" table
CREATE TABLE "public"."organizations" (
  "id" bigint NOT NULL,
  "updated_at" timestamptz NULL,
  "name" text NOT NULL,
  "is_default" boolean NOT NULL DEFAULT false,
  "workspace_id" bigint NOT NULL,
  "cc_addresses" jsonb NOT NULL DEFAULT '[]',
  "subdomain" text NOT NULL,
  "auto_opt_in" boolean NOT NULL DEFAULT false,
  "auto_opt_in_template_id" bigint NULL,
  PRIMARY KEY ("id")
);
-- Create "sns_topics" table
CREATE TABLE "public"."sns_topics" (
  "id" bigint NOT NULL,
  "updated_at" timestamptz NULL,
  "region" text NOT NULL,
  "arn" text NULL,
  "status" text NULL,
  PRIMARY KEY ("id")
);
-- Create "team_users" table
CREATE TABLE "public"."team_users" (
  "id" bigint NOT NULL,
  "updated_at" timestamptz NULL,
  "status" "public"."team_user_status" NOT NULL DEFAULT 'PENDING',
  "team_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "organization_id" bigint NOT NULL,
  "workspace_id" bigint NOT NULL,
  "last_login_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "teams" table
CREATE TABLE "public"."teams" (
  "id" bigint NOT NULL,
  "updated_at" timestamptz NULL,
  "name" text NULL,
  "is_active" boolean NULL,
  "permissions" jsonb NOT NULL DEFAULT '[]',
  "organization_id" bigint NOT NULL,
  "workspace_id" bigint NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "templates" table
CREATE TABLE "public"."templates" (
  "id" bigint NOT NULL,
  "updated_at" timestamptz NULL,
  "name" text NULL,
  "content_engine" text NOT NULL,
  "is_transactional" boolean NOT NULL DEFAULT true,
  "subject" text NULL,
  "alt_subject" text NULL,
  "content" text NULL,
  "text_content" text NULL,
  "is_opt_in" boolean NOT NULL DEFAULT false,
  "organization_id" bigint NOT NULL,
  "workspace_id" bigint NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "users" table
CREATE TABLE "public"."users" (
  "id" bigint NOT NULL,
  "updated_at" timestamptz NULL,
  "first_name" text NULL,
  "last_name" text NULL,
  "email" text NOT NULL,
  "email_verified" boolean NULL,
  "identity_provider" "public"."identity_provider" NOT NULL DEFAULT 'LOCAL',
  PRIMARY KEY ("id")
);
-- Create index "idx_email" to table: "users"
CREATE UNIQUE INDEX "idx_email" ON "public"."users" ("email") WHERE (email IS NOT NULL);
-- Create "workspaces" table
CREATE TABLE "public"."workspaces" (
  "id" bigint NOT NULL,
  "updated_at" timestamptz NULL,
  "name" text NULL,
  "owner" bigint NULL,
  "address" jsonb NOT NULL DEFAULT '{}',
  PRIMARY KEY ("id")
);
