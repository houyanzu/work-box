ALTER TABLE `box_monitor_chain_record`
    ADD COLUMN `chain_db_id` int(11) UNSIGNED NOT NULL DEFAULT 1 AFTER `id`
;

ALTER TABLE `box_transfer_records`
    ADD COLUMN `chain_db_id` int(11) UNSIGNED NOT NULL DEFAULT 1 AFTER `id`
;

ALTER TABLE `box_lock_transfer_details`
    ADD COLUMN `chain_db_id` int(11) UNSIGNED NOT NULL DEFAULT 1 AFTER `id`
;

ALTER TABLE `box_transfer_details`
    ADD COLUMN `chain_db_id` int(11) UNSIGNED NOT NULL DEFAULT 1 AFTER `id`
;

ALTER TABLE `box_user_keys_balance`
    ADD COLUMN `chain_db_id` int(11) UNSIGNED NOT NULL DEFAULT 1 AFTER `id`,
    ADD COLUMN `status` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0-正常，1-待转eth，2-转eth中' AFTER `balance`,
    ADD COLUMN `collect_status` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0-正常，1-归集中' AFTER `status`,
    ADD COLUMN `module_id` int(11) UNSIGNED NOT NULL DEFAULT 0 AFTER `collect_status`
;

ALTER TABLE `box_tokens`
    ADD COLUMN `chain_db_id` int(11) UNSIGNED NOT NULL DEFAULT 1 AFTER `id`
;

ALTER TABLE `box_uk_collect_record`
    ADD COLUMN `chain_db_id` int(11) UNSIGNED NOT NULL DEFAULT 1 AFTER `id`
;