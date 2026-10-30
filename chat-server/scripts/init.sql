SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

INSERT INTO `role` (`name`, `code`, `description`, `status`, `created_at`, `updated_at`) VALUES
  ('超级管理员', 'super_admin', '拥有系统全部权限', 1, NOW(), NOW());

INSERT INTO `permission` (`name`, `code`, `description`, `type`, `parent_id`, `path`, `method`, `sort`, `status`, `icon`, `component`, `hide_in_menu`, `created_at`, `updated_at`) VALUES
  ('组织权限', 'org', '', 1, 0, '/org', '', 100, 1, 'Lock', '', 0, NOW(), NOW()),
  ('系统运维', 'ops', '', 1, 0, '/ops', '', 200, 1, 'Setting', '', 0, NOW(), NOW()),
  ('用户管理', 'user', '', 1, 0, '/user', '', 300, 1, 'User', '', 0, NOW(), NOW()),
  ('沟通协作', 'im', '', 1, 0, '/im', '', 400, 1, 'Headset', '', 0, NOW(), NOW()),
  ('社区治理', 'chat', '', 1, 0, '/chat', '', 500, 1, 'Monitor', '', 0, NOW(), NOW()),
  ('帖子管理', 'post', '', 1, 0, '/post', '', 600, 1, 'Document', '', 0, NOW(), NOW()),
  ('广告管理', 'ad', '', 1, 0, '/ad', '', 700, 1, 'Goods', '', 0, NOW(), NOW());

SET @pid_org = (SELECT `id` FROM `permission` WHERE `code` = 'org' AND `type` = 1 AND `parent_id` = 0 LIMIT 1);
SET @pid_ops = (SELECT `id` FROM `permission` WHERE `code` = 'ops' AND `type` = 1 AND `parent_id` = 0 LIMIT 1);
SET @pid_user = (SELECT `id` FROM `permission` WHERE `code` = 'user' AND `type` = 1 AND `parent_id` = 0 LIMIT 1);
SET @pid_im = (SELECT `id` FROM `permission` WHERE `code` = 'im' AND `type` = 1 AND `parent_id` = 0 LIMIT 1);
SET @pid_chat = (SELECT `id` FROM `permission` WHERE `code` = 'chat' AND `type` = 1 AND `parent_id` = 0 LIMIT 1);
SET @pid_post = (SELECT `id` FROM `permission` WHERE `code` = 'post' AND `type` = 1 AND `parent_id` = 0 LIMIT 1);
SET @pid_ad = (SELECT `id` FROM `permission` WHERE `code` = 'ad' AND `type` = 1 AND `parent_id` = 0 LIMIT 1);

INSERT INTO `permission` (`name`, `code`, `description`, `type`, `parent_id`, `path`, `method`, `sort`, `status`, `icon`, `component`, `hide_in_menu`, `created_at`, `updated_at`) VALUES
  ('员工管理', 'org:employee', '', 1, @pid_org, '/org/employee', '', 10, 1, 'Avatar', 'EmployeeList', 0, NOW(), NOW()),
  ('部门管理', 'org:department', '', 1, @pid_org, '/org/department', '', 20, 1, 'OfficeBuilding', 'DepartmentList', 0, NOW(), NOW()),
  ('角色管理', 'org:role', '', 1, @pid_org, '/org/role', '', 30, 1, 'Lock', 'RoleList', 0, NOW(), NOW()),
  ('权限管理', 'org:permission', '', 1, @pid_org, '/org/permission', '', 40, 1, 'Key', 'PermissionList', 0, NOW(), NOW()),
  ('行政区划', 'ops:region', '', 1, @pid_ops, '/ops/region', '', 10, 1, 'MapLocation', 'RegionTree', 0, NOW(), NOW()),
  ('操作审计', 'ops:audit_log', '', 1, @pid_ops, '/ops/audit-log', '', 20, 1, 'Document', 'AuditLogList', 0, NOW(), NOW()),
  ('应用发布', 'ops:app_release', '', 1, @pid_ops, '/ops/app-release', '', 30, 1, 'Download', 'AppReleaseList', 0, NOW(), NOW()),
  ('屏蔽词', 'ops:sensitive_word', '', 1, @pid_ops, '/ops/sensitive-word', '', 40, 1, 'Filter', 'SensitiveWordList', 0, NOW(), NOW()),
  ('用户账号', 'user:account', '', 1, @pid_user, '/user/account', '', 10, 1, 'User', 'UserList', 0, NOW(), NOW()),
  ('用户反馈', 'user:feedback', '', 1, @pid_user, '/user/feedback', '', 20, 1, 'ChatLineSquare', 'FeedbackList', 0, NOW(), NOW()),
  ('举报管理', 'user:report', '', 1, @pid_user, '/user/report', '', 30, 1, 'Warning', 'ReportList', 0, NOW(), NOW()),
  ('员工会话', 'im:chat', '', 1, @pid_im, '/im/chat', '', 10, 1, 'ChatDotRound', 'StaffChat', 0, NOW(), NOW()),
  ('官方客服', 'im:support', '', 1, @pid_im, '/im/support', '', 20, 1, 'Headset', 'SupportDesk', 0, NOW(), NOW()),
  ('客服总览', 'im:support_overview', '', 1, @pid_im, '/im/support-overview', '', 30, 1, 'DataBoard', 'SupportOverview', 0, NOW(), NOW()),
  ('地理围栏', 'chat:room', '', 1, @pid_chat, '/chat/room', '', 10, 1, 'Location', 'RoomList', 0, NOW(), NOW()),
  ('聊天消息', 'chat:message', '', 1, @pid_chat, '/chat/message', '', 20, 1, 'Message', 'MessageList', 0, NOW(), NOW()),
  ('私信会话', 'chat:conversation', '', 1, @pid_chat, '/chat/conversation', '', 30, 1, 'ChatLineSquare', 'ConversationList', 0, NOW(), NOW()),
  ('帖子列表', 'post:post', '', 1, @pid_post, '/post/list', '', 10, 1, 'Document', 'PostList', 0, NOW(), NOW()),
  ('广告计划', 'ad:campaign', '', 1, @pid_ad, '/ad/campaign', '', 10, 1, 'Goods', 'AdCampaignList', 0, NOW(), NOW()),
  ('投放设置', 'ad:setting', '', 1, @pid_ad, '/ad/setting', '', 20, 1, 'Setting', 'AdSetting', 0, NOW(), NOW());

SET @pid_employee = (SELECT `id` FROM `permission` WHERE `code` = 'org:employee' AND `type` = 1 LIMIT 1);
SET @pid_department = (SELECT `id` FROM `permission` WHERE `code` = 'org:department' AND `type` = 1 LIMIT 1);
SET @pid_role = (SELECT `id` FROM `permission` WHERE `code` = 'org:role' AND `type` = 1 LIMIT 1);
SET @pid_perm = (SELECT `id` FROM `permission` WHERE `code` = 'org:permission' AND `type` = 1 LIMIT 1);
SET @pid_region = (SELECT `id` FROM `permission` WHERE `code` = 'ops:region' AND `type` = 1 LIMIT 1);
SET @pid_audit_log = (SELECT `id` FROM `permission` WHERE `code` = 'ops:audit_log' AND `type` = 1 LIMIT 1);
SET @pid_app_release = (SELECT `id` FROM `permission` WHERE `code` = 'ops:app_release' AND `type` = 1 LIMIT 1);
SET @pid_sensitive_word = (SELECT `id` FROM `permission` WHERE `code` = 'ops:sensitive_word' AND `type` = 1 LIMIT 1);
SET @pid_user_account = (SELECT `id` FROM `permission` WHERE `code` = 'user:account' AND `type` = 1 LIMIT 1);
SET @pid_user_feedback = (SELECT `id` FROM `permission` WHERE `code` = 'user:feedback' AND `type` = 1 LIMIT 1);
SET @pid_user_report = (SELECT `id` FROM `permission` WHERE `code` = 'user:report' AND `type` = 1 LIMIT 1);
SET @pid_im_chat = (SELECT `id` FROM `permission` WHERE `code` = 'im:chat' AND `type` = 1 LIMIT 1);
SET @pid_im_support = (SELECT `id` FROM `permission` WHERE `code` = 'im:support' AND `type` = 1 LIMIT 1);
SET @pid_im_support_ov = (SELECT `id` FROM `permission` WHERE `code` = 'im:support_overview' AND `type` = 1 LIMIT 1);
SET @pid_chat_room = (SELECT `id` FROM `permission` WHERE `code` = 'chat:room' AND `type` = 1 LIMIT 1);
SET @pid_chat_message = (SELECT `id` FROM `permission` WHERE `code` = 'chat:message' AND `type` = 1 LIMIT 1);
SET @pid_chat_conv = (SELECT `id` FROM `permission` WHERE `code` = 'chat:conversation' AND `type` = 1 LIMIT 1);
SET @pid_post_post = (SELECT `id` FROM `permission` WHERE `code` = 'post:post' AND `type` = 1 LIMIT 1);
SET @pid_ad_campaign = (SELECT `id` FROM `permission` WHERE `code` = 'ad:campaign' AND `type` = 1 LIMIT 1);
SET @pid_ad_setting = (SELECT `id` FROM `permission` WHERE `code` = 'ad:setting' AND `type` = 1 LIMIT 1);

INSERT INTO `permission` (`name`, `code`, `description`, `type`, `parent_id`, `path`, `method`, `sort`, `status`, `created_at`, `updated_at`) VALUES
  ('新增员工', 'org:employee:create', '', 2, @pid_employee, '', '', 10, 1, NOW(), NOW()),
  ('编辑员工', 'org:employee:update', '', 2, @pid_employee, '', '', 20, 1, NOW(), NOW()),
  ('员工状态', 'org:employee:status', '', 2, @pid_employee, '', '', 30, 1, NOW(), NOW()),
  ('分配角色', 'org:employee:assign_role', '', 2, @pid_employee, '', '', 40, 1, NOW(), NOW()),
  ('部门详情', 'org:department:detail', '', 2, @pid_department, '', '', 10, 1, NOW(), NOW()),
  ('新增部门', 'org:department:create', '', 2, @pid_department, '', '', 20, 1, NOW(), NOW()),
  ('编辑部门', 'org:department:update', '', 2, @pid_department, '', '', 30, 1, NOW(), NOW()),
  ('部门状态', 'org:department:status', '', 2, @pid_department, '', '', 40, 1, NOW(), NOW()),
  ('新增角色', 'org:role:create', '', 2, @pid_role, '', '', 10, 1, NOW(), NOW()),
  ('编辑角色', 'org:role:update', '', 2, @pid_role, '', '', 20, 1, NOW(), NOW()),
  ('删除角色', 'org:role:delete', '', 2, @pid_role, '', '', 30, 1, NOW(), NOW()),
  ('分配权限', 'org:role:assign_perm', '', 2, @pid_role, '', '', 40, 1, NOW(), NOW()),
  ('新增权限', 'org:permission:create', '', 2, @pid_perm, '', '', 10, 1, NOW(), NOW()),
  ('编辑权限', 'org:permission:update', '', 2, @pid_perm, '', '', 20, 1, NOW(), NOW()),
  ('删除权限', 'org:permission:delete', '', 2, @pid_perm, '', '', 30, 1, NOW(), NOW()),
  ('区划详情', 'ops:region:detail', '', 2, @pid_region, '', '', 10, 1, NOW(), NOW()),
  ('审计详情', 'ops:audit_log:detail', '', 2, @pid_audit_log, '', '', 10, 1, NOW(), NOW()),
  ('上传安装包', 'ops:app_release:upload', '', 2, @pid_app_release, '', '', 10, 1, NOW(), NOW()),
  ('登记网页版', 'ops:app_release:publish', '', 2, @pid_app_release, '', '', 20, 1, NOW(), NOW()),
  ('设为当前', 'ops:app_release:enable', '', 2, @pid_app_release, '', '', 30, 1, NOW(), NOW()),
  ('下架当前', 'ops:app_release:disable', '', 2, @pid_app_release, '', '', 40, 1, NOW(), NOW()),
  ('更新策略', 'ops:app_release:policy', '', 2, @pid_app_release, '', '', 50, 1, NOW(), NOW()),
  ('删除版本', 'ops:app_release:delete', '', 2, @pid_app_release, '', '', 60, 1, NOW(), NOW()),
  ('媒体上传', 'ops:upload', '', 2, @pid_ops, '', '', 50, 1, NOW(), NOW()),
  ('用户详情', 'user:account:detail', '', 2, @pid_user_account, '', '', 10, 1, NOW(), NOW()),
  ('新增用户', 'user:account:create', '', 2, @pid_user_account, '', '', 20, 1, NOW(), NOW()),
  ('编辑用户', 'user:account:update', '', 2, @pid_user_account, '', '', 30, 1, NOW(), NOW()),
  ('用户状态', 'user:account:status', '', 2, @pid_user_account, '', '', 40, 1, NOW(), NOW()),
  ('用户禁言', 'user:account:mute', '', 2, @pid_user_account, '', '', 50, 1, NOW(), NOW()),
  ('注销用户', 'user:account:deactivate', '', 2, @pid_user_account, '', '', 60, 1, NOW(), NOW()),
  ('反馈详情', 'user:feedback:detail', '', 2, @pid_user_feedback, '', '', 10, 1, NOW(), NOW()),
  ('处理反馈', 'user:feedback:status', '', 2, @pid_user_feedback, '', '', 20, 1, NOW(), NOW()),
  ('举报详情', 'user:report:detail', '', 2, @pid_user_report, '', '', 10, 1, NOW(), NOW()),
  ('处理举报', 'user:report:handle', '', 2, @pid_user_report, '', '', 20, 1, NOW(), NOW()),
  ('客服接入', 'im:support:accept', '', 2, @pid_im_support, '', '', 10, 1, NOW(), NOW()),
  ('客服转接', 'im:support:transfer', '', 2, @pid_im_support, '', '', 20, 1, NOW(), NOW()),
  ('客服结案', 'im:support:close', '', 2, @pid_im_support, '', '', 30, 1, NOW(), NOW()),
  ('客服回复', 'im:support:send', '', 2, @pid_im_support, '', '', 40, 1, NOW(), NOW()),
  ('客服配置', 'im:support_overview:config', '', 2, @pid_im_support_ov, '', '', 10, 1, NOW(), NOW()),
  ('打开私聊', 'im:chat:open', '', 2, @pid_im_chat, '', '', 10, 1, NOW(), NOW()),
  ('发送消息', 'im:chat:send', '', 2, @pid_im_chat, '', '', 20, 1, NOW(), NOW()),
  ('标记已读', 'im:chat:read', '', 2, @pid_im_chat, '', '', 30, 1, NOW(), NOW()),
  ('撤回消息', 'im:chat:recall', '', 2, @pid_im_chat, '', '', 40, 1, NOW(), NOW()),
  ('新增围栏', 'chat:room:create', '', 2, @pid_chat_room, '', '', 10, 1, NOW(), NOW()),
  ('编辑围栏', 'chat:room:update', '', 2, @pid_chat_room, '', '', 20, 1, NOW(), NOW()),
  ('审核围栏', 'chat:room:audit', '', 2, @pid_chat_room, '', '', 30, 1, NOW(), NOW()),
  ('围栏状态', 'chat:room:status', '', 2, @pid_chat_room, '', '', 40, 1, NOW(), NOW()),
  ('消息上下文', 'chat:message:context', '', 2, @pid_chat_message, '', '', 10, 1, NOW(), NOW()),
  ('删除消息', 'chat:message:delete', '', 2, @pid_chat_message, '', '', 20, 1, NOW(), NOW()),
  ('帖子详情', 'post:post:detail', '', 2, @pid_post_post, '', '', 10, 1, NOW(), NOW()),
  ('帖子状态', 'post:post:status', '', 2, @pid_post_post, '', '', 20, 1, NOW(), NOW()),
  ('新增屏蔽词', 'ops:sensitive_word:create', '', 2, @pid_sensitive_word, '', '', 10, 1, NOW(), NOW()),
  ('编辑屏蔽词', 'ops:sensitive_word:update', '', 2, @pid_sensitive_word, '', '', 20, 1, NOW(), NOW()),
  ('屏蔽词状态', 'ops:sensitive_word:status', '', 2, @pid_sensitive_word, '', '', 30, 1, NOW(), NOW()),
  ('删除屏蔽词', 'ops:sensitive_word:delete', '', 2, @pid_sensitive_word, '', '', 40, 1, NOW(), NOW()),
  ('导入屏蔽词', 'ops:sensitive_word:import', '', 2, @pid_sensitive_word, '', '', 50, 1, NOW(), NOW()),
  ('保存广告', 'ad:campaign:update', '', 2, @pid_ad_campaign, '', '', 10, 1, NOW(), NOW()),
  ('广告状态', 'ad:campaign:status', '', 2, @pid_ad_campaign, '', '', 20, 1, NOW(), NOW()),
  ('复制广告', 'ad:campaign:copy', '', 2, @pid_ad_campaign, '', '', 30, 1, NOW(), NOW()),
  ('批量暂停', 'ad:campaign:pause', '', 2, @pid_ad_campaign, '', '', 40, 1, NOW(), NOW()),
  ('广告日报', 'ad:campaign:stats', '', 2, @pid_ad_campaign, '', '', 50, 1, NOW(), NOW()),
  ('审核广告', 'ad:campaign:audit', '', 2, @pid_ad_campaign, '', '', 60, 1, NOW(), NOW()),
  ('保存投放设置', 'ad:setting:update', '', 2, @pid_ad_setting, '', '', 10, 1, NOW(), NOW());

INSERT INTO `permission` (`name`, `code`, `description`, `type`, `parent_id`, `path`, `method`, `sort`, `status`, `created_at`, `updated_at`) VALUES
  ('员工列表', 'org:employee', '', 3, @pid_employee, '/api/admin/employee/list', 'GET', 10, 1, NOW(), NOW()),
  ('员工新增', 'org:employee:create', '', 3, @pid_employee, '/api/admin/employee/create', 'POST', 20, 1, NOW(), NOW()),
  ('员工编辑', 'org:employee:update', '', 3, @pid_employee, '/api/admin/employee/update', 'POST', 30, 1, NOW(), NOW()),
  ('员工状态', 'org:employee:status', '', 3, @pid_employee, '/api/admin/employee/update/status', 'POST', 40, 1, NOW(), NOW()),
  ('员工角色选项', 'org:employee:assign_role', '', 3, @pid_employee, '/api/admin/employee/role-options', 'GET', 50, 1, NOW(), NOW()),
  ('分配员工角色', 'org:employee:assign_role', '', 3, @pid_employee, '/api/admin/employee/assign/roles', 'POST', 60, 1, NOW(), NOW()),
  ('员工登录记录', 'org:employee', '', 3, @pid_employee, '/api/admin/employee/login-records', 'GET', 70, 1, NOW(), NOW()),
  ('部门列表', 'org:department', '', 3, @pid_department, '/api/admin/department/list', 'GET', 10, 1, NOW(), NOW()),
  ('部门新增', 'org:department:create', '', 3, @pid_department, '/api/admin/department/create', 'POST', 20, 1, NOW(), NOW()),
  ('部门编辑', 'org:department:update', '', 3, @pid_department, '/api/admin/department/update', 'POST', 30, 1, NOW(), NOW()),
  ('部门状态', 'org:department:status', '', 3, @pid_department, '/api/admin/department/update/status', 'POST', 40, 1, NOW(), NOW()),
  ('角色列表', 'org:role', '', 3, @pid_role, '/api/admin/role/list', 'GET', 10, 1, NOW(), NOW()),
  ('角色新增', 'org:role:create', '', 3, @pid_role, '/api/admin/role/create', 'POST', 30, 1, NOW(), NOW()),
  ('角色编辑', 'org:role:update', '', 3, @pid_role, '/api/admin/role/update', 'POST', 40, 1, NOW(), NOW()),
  ('角色删除', 'org:role:delete', '', 3, @pid_role, '/api/admin/role/delete', 'POST', 50, 1, NOW(), NOW()),
  ('角色权限查询', 'org:role', '', 3, @pid_role, '/api/admin/role/permissions', 'GET', 60, 1, NOW(), NOW()),
  ('分配角色权限', 'org:role:assign_perm', '', 3, @pid_role, '/api/admin/role/assign/permissions', 'POST', 70, 1, NOW(), NOW()),
  ('权限树', 'org:permission', '', 3, @pid_perm, '/api/admin/permission/tree', 'GET', 30, 1, NOW(), NOW()),
  ('权限新增', 'org:permission:create', '', 3, @pid_perm, '/api/admin/permission/create', 'POST', 40, 1, NOW(), NOW()),
  ('权限编辑', 'org:permission:update', '', 3, @pid_perm, '/api/admin/permission/update', 'POST', 50, 1, NOW(), NOW()),
  ('权限删除', 'org:permission:delete', '', 3, @pid_perm, '/api/admin/permission/delete', 'POST', 60, 1, NOW(), NOW()),
  ('区划子级', 'ops:region', '', 3, @pid_region, '/api/admin/region/children', 'GET', 10, 1, NOW(), NOW()),
  ('区划搜索', 'ops:region', '', 3, @pid_region, '/api/admin/region/search', 'GET', 20, 1, NOW(), NOW()),
  ('区划路径', 'ops:region', '', 3, @pid_region, '/api/admin/region/trail', 'GET', 25, 1, NOW(), NOW()),
  ('区划详情', 'ops:region:detail', '', 3, @pid_region, '/api/admin/region/detail', 'GET', 30, 1, NOW(), NOW()),
  ('审计列表', 'ops:audit_log', '', 3, @pid_audit_log, '/api/admin/audit/list', 'GET', 10, 1, NOW(), NOW()),
  ('审计详情', 'ops:audit_log:detail', '', 3, @pid_audit_log, '/api/admin/audit/detail', 'GET', 20, 1, NOW(), NOW()),
  ('审计筛选项', 'ops:audit_log', '', 3, @pid_audit_log, '/api/admin/audit/options', 'GET', 30, 1, NOW(), NOW()),
  ('审计操作人搜索', 'ops:audit_log', '', 3, @pid_audit_log, '/api/admin/audit/employees', 'GET', 40, 1, NOW(), NOW()),
  ('应用版本列表', 'ops:app_release', '', 3, @pid_app_release, '/api/admin/app/release/list', 'GET', 10, 1, NOW(), NOW()),
  ('上传安装包', 'ops:app_release:upload', '', 3, @pid_app_release, '/api/admin/app/release/upload', 'POST', 20, 1, NOW(), NOW()),
  ('登记网页版', 'ops:app_release:publish', '', 3, @pid_app_release, '/api/admin/app/release/publish-meta', 'POST', 30, 1, NOW(), NOW()),
  ('设为当前包', 'ops:app_release:enable', '', 3, @pid_app_release, '/api/admin/app/release/set-current', 'POST', 40, 1, NOW(), NOW()),
  ('下架当前包', 'ops:app_release:disable', '', 3, @pid_app_release, '/api/admin/app/release/unset-current', 'POST', 50, 1, NOW(), NOW()),
  ('更新策略列表', 'ops:app_release', '', 3, @pid_app_release, '/api/admin/app/release/policy/list', 'GET', 60, 1, NOW(), NOW()),
  ('保存更新策略', 'ops:app_release:policy', '', 3, @pid_app_release, '/api/admin/app/release/policy/save', 'POST', 70, 1, NOW(), NOW()),
  ('删除版本', 'ops:app_release:delete', '', 3, @pid_app_release, '/api/admin/app/release/delete', 'POST', 80, 1, NOW(), NOW()),
  ('图片上传', 'ops:upload', '', 3, @pid_ops, '/api/admin/upload/image', 'POST', 10, 1, NOW(), NOW()),
  ('视频上传', 'ops:upload', '', 3, @pid_ops, '/api/admin/upload/video', 'POST', 20, 1, NOW(), NOW()),
  ('文件上传', 'ops:upload', '', 3, @pid_ops, '/api/admin/upload/file', 'POST', 30, 1, NOW(), NOW()),
  ('屏蔽词列表', 'ops:sensitive_word', '', 3, @pid_sensitive_word, '/api/admin/sensitive-word/list', 'GET', 10, 1, NOW(), NOW()),
  ('屏蔽词新增', 'ops:sensitive_word:create', '', 3, @pid_sensitive_word, '/api/admin/sensitive-word/create', 'POST', 20, 1, NOW(), NOW()),
  ('屏蔽词编辑', 'ops:sensitive_word:update', '', 3, @pid_sensitive_word, '/api/admin/sensitive-word/update', 'POST', 30, 1, NOW(), NOW()),
  ('屏蔽词状态', 'ops:sensitive_word:status', '', 3, @pid_sensitive_word, '/api/admin/sensitive-word/update/status', 'POST', 40, 1, NOW(), NOW()),
  ('屏蔽词删除', 'ops:sensitive_word:delete', '', 3, @pid_sensitive_word, '/api/admin/sensitive-word/delete', 'POST', 50, 1, NOW(), NOW()),
  ('屏蔽词导入', 'ops:sensitive_word:import', '', 3, @pid_sensitive_word, '/api/admin/sensitive-word/import', 'POST', 60, 1, NOW(), NOW()),
  ('屏蔽词批量删除', 'ops:sensitive_word:delete', '', 3, @pid_sensitive_word, '/api/admin/sensitive-word/batch/delete', 'POST', 70, 1, NOW(), NOW()),
  ('屏蔽词批量状态', 'ops:sensitive_word:status', '', 3, @pid_sensitive_word, '/api/admin/sensitive-word/batch/status', 'POST', 80, 1, NOW(), NOW()),
  ('用户列表', 'user:account', '', 3, @pid_user_account, '/api/admin/user/list', 'GET', 10, 1, NOW(), NOW()),
  ('用户详情', 'user:account:detail', '', 3, @pid_user_account, '/api/admin/user/detail', 'GET', 20, 1, NOW(), NOW()),
  ('用户新增', 'user:account:create', '', 3, @pid_user_account, '/api/admin/user/create', 'POST', 30, 1, NOW(), NOW()),
  ('用户编辑', 'user:account:update', '', 3, @pid_user_account, '/api/admin/user/update', 'POST', 40, 1, NOW(), NOW()),
  ('用户状态', 'user:account:status', '', 3, @pid_user_account, '/api/admin/user/update/status', 'POST', 50, 1, NOW(), NOW()),
  ('用户禁言', 'user:account:mute', '', 3, @pid_user_account, '/api/admin/user/update/mute', 'POST', 60, 1, NOW(), NOW()),
  ('用户注销', 'user:account:deactivate', '', 3, @pid_user_account, '/api/admin/user/deactivate', 'POST', 70, 1, NOW(), NOW()),
  ('用户批量禁言', 'user:account:mute', '', 3, @pid_user_account, '/api/admin/user/batch/mute', 'POST', 80, 1, NOW(), NOW()),
  ('用户批量状态', 'user:account:status', '', 3, @pid_user_account, '/api/admin/user/batch/status', 'POST', 90, 1, NOW(), NOW()),
  ('用户处罚记录', 'user:account:detail', '', 3, @pid_user_account, '/api/admin/user/punishments', 'GET', 100, 1, NOW(), NOW()),
  ('用户登录记录', 'user:account:detail', '', 3, @pid_user_account, '/api/admin/user/login-records', 'GET', 110, 1, NOW(), NOW()),
  ('用户设备', 'user:account:detail', '', 3, @pid_user_account, '/api/admin/user/devices', 'GET', 120, 1, NOW(), NOW()),
  ('查看完整手机号', 'user:account:detail', '', 3, @pid_user_account, '/api/admin/user/mobile/reveal', 'POST', 130, 1, NOW(), NOW()),
  ('反馈列表', 'user:feedback', '', 3, @pid_user_feedback, '/api/admin/feedback/list', 'GET', 10, 1, NOW(), NOW()),
  ('反馈详情', 'user:feedback:detail', '', 3, @pid_user_feedback, '/api/admin/feedback/detail', 'GET', 20, 1, NOW(), NOW()),
  ('反馈状态更新', 'user:feedback:status', '', 3, @pid_user_feedback, '/api/admin/feedback/update/status', 'POST', 30, 1, NOW(), NOW()),
  ('反馈批量关闭', 'user:feedback:status', '', 3, @pid_user_feedback, '/api/admin/feedback/batch/close', 'POST', 40, 1, NOW(), NOW()),
  ('举报列表', 'user:report', '', 3, @pid_user_report, '/api/admin/report/list', 'GET', 10, 1, NOW(), NOW()),
  ('举报详情', 'user:report:detail', '', 3, @pid_user_report, '/api/admin/report/detail', 'GET', 20, 1, NOW(), NOW()),
  ('举报处理', 'user:report:handle', '', 3, @pid_user_report, '/api/admin/report/handle', 'POST', 30, 1, NOW(), NOW()),
  ('会话通讯录', 'im:chat', '', 3, @pid_im_chat, '/api/admin/im/contacts', 'GET', 10, 1, NOW(), NOW()),
  ('打开私聊', 'im:chat:open', '', 3, @pid_im_chat, '/api/admin/im/conversations/dm', 'POST', 20, 1, NOW(), NOW()),
  ('会话列表', 'im:chat', '', 3, @pid_im_chat, '/api/admin/im/conversations', 'GET', 30, 1, NOW(), NOW()),
  ('会话消息', 'im:chat', '', 3, @pid_im_chat, '/api/admin/im/messages', 'GET', 40, 1, NOW(), NOW()),
  ('发送消息', 'im:chat:send', '', 3, @pid_im_chat, '/api/admin/im/messages', 'POST', 50, 1, NOW(), NOW()),
  ('标记已读', 'im:chat:read', '', 3, @pid_im_chat, '/api/admin/im/read', 'POST', 60, 1, NOW(), NOW()),
  ('撤回消息', 'im:chat:recall', '', 3, @pid_im_chat, '/api/admin/im/recall', 'POST', 70, 1, NOW(), NOW()),
  ('客服排队', 'im:support', '', 3, @pid_im_support, '/api/admin/support/queue', 'GET', 10, 1, NOW(), NOW()),
  ('我的客服会话', 'im:support', '', 3, @pid_im_support, '/api/admin/support/mine', 'GET', 20, 1, NOW(), NOW()),
  ('已结案会话', 'im:support', '', 3, @pid_im_support, '/api/admin/support/closed', 'GET', 30, 1, NOW(), NOW()),
  ('客服接入', 'im:support:accept', '', 3, @pid_im_support, '/api/admin/support/accept', 'POST', 40, 1, NOW(), NOW()),
  ('客服转接', 'im:support:transfer', '', 3, @pid_im_support, '/api/admin/support/transfer', 'POST', 50, 1, NOW(), NOW()),
  ('客服结案', 'im:support:close', '', 3, @pid_im_support, '/api/admin/support/close', 'POST', 60, 1, NOW(), NOW()),
  ('客服消息列表', 'im:support', '', 3, @pid_im_support, '/api/admin/support/messages', 'GET', 70, 1, NOW(), NOW()),
  ('客服回复', 'im:support:send', '', 3, @pid_im_support, '/api/admin/support/messages', 'POST', 80, 1, NOW(), NOW()),
  ('客服已读', 'im:support:send', '', 3, @pid_im_support, '/api/admin/support/read', 'POST', 90, 1, NOW(), NOW()),
  ('客服快捷回复', 'im:support', '', 3, @pid_im_support, '/api/admin/support/quick-replies', 'GET', 100, 1, NOW(), NOW()),
  ('用户来访历史', 'im:support', '', 3, @pid_im_support, '/api/admin/support/history', 'GET', 110, 1, NOW(), NOW()),
  ('客服坐席列表', 'im:support_overview', '', 3, @pid_im_support, '/api/admin/support/agents', 'GET', 120, 1, NOW(), NOW()),
  ('全量进行中会话', 'im:support_overview', '', 3, @pid_im_support_ov, '/api/admin/support/active-all', 'GET', 10, 1, NOW(), NOW()),
  ('全量已结案会话', 'im:support_overview', '', 3, @pid_im_support_ov, '/api/admin/support/closed-all', 'GET', 20, 1, NOW(), NOW()),
  ('客服总览摘要', 'im:support_overview', '', 3, @pid_im_support_ov, '/api/admin/support/overview/summary', 'GET', 30, 1, NOW(), NOW()),
  ('客服配置读', 'im:support_overview', '', 3, @pid_im_support_ov, '/api/admin/support/config', 'GET', 40, 1, NOW(), NOW()),
  ('客服配置写', 'im:support_overview:config', '', 3, @pid_im_support_ov, '/api/admin/support/config', 'POST', 50, 1, NOW(), NOW()),
  ('客服会话详情', 'im:support_overview', '', 3, @pid_im_support_ov, '/api/admin/support/session', 'GET', 60, 1, NOW(), NOW()),
  ('用户客服记录', 'im:support_overview', '', 3, @pid_im_support_ov, '/api/admin/support/by-user', 'GET', 70, 1, NOW(), NOW()),
  ('围栏列表', 'chat:room', '', 3, @pid_chat_room, '/api/admin/room/list', 'GET', 10, 1, NOW(), NOW()),
  ('围栏详情', 'chat:room', '', 3, @pid_chat_room, '/api/admin/room/detail', 'GET', 20, 1, NOW(), NOW()),
  ('围栏邻近', 'chat:room', '', 3, @pid_chat_room, '/api/admin/room/neighbors', 'GET', 30, 1, NOW(), NOW()),
  ('围栏全景', 'chat:room', '', 3, @pid_chat_room, '/api/admin/room/map', 'GET', 40, 1, NOW(), NOW()),
  ('围栏新增', 'chat:room:create', '', 3, @pid_chat_room, '/api/admin/room/create', 'POST', 50, 1, NOW(), NOW()),
  ('围栏编辑', 'chat:room:update', '', 3, @pid_chat_room, '/api/admin/room/update', 'POST', 60, 1, NOW(), NOW()),
  ('围栏审核', 'chat:room:audit', '', 3, @pid_chat_room, '/api/admin/room/audit', 'POST', 70, 1, NOW(), NOW()),
  ('围栏批量审核', 'chat:room:audit', '', 3, @pid_chat_room, '/api/admin/room/batch/audit', 'POST', 75, 1, NOW(), NOW()),
  ('围栏状态', 'chat:room:status', '', 3, @pid_chat_room, '/api/admin/room/update/status', 'POST', 80, 1, NOW(), NOW()),
  ('消息列表', 'chat:message', '', 3, @pid_chat_message, '/api/admin/message/list', 'GET', 10, 1, NOW(), NOW()),
  ('消息上下文', 'chat:message:context', '', 3, @pid_chat_message, '/api/admin/message/context', 'GET', 20, 1, NOW(), NOW()),
  ('删除消息', 'chat:message:delete', '', 3, @pid_chat_message, '/api/admin/message/batch/delete', 'POST', 30, 1, NOW(), NOW()),
  ('私信列表', 'chat:conversation', '', 3, @pid_chat_conv, '/api/admin/conversation/list', 'GET', 10, 1, NOW(), NOW()),
  ('帖子列表', 'post:post', '', 3, @pid_post_post, '/api/admin/post/list', 'GET', 10, 1, NOW(), NOW()),
  ('帖子详情', 'post:post:detail', '', 3, @pid_post_post, '/api/admin/post/detail', 'GET', 20, 1, NOW(), NOW()),
  ('帖子全景', 'post:post', '', 3, @pid_post_post, '/api/admin/post/map', 'GET', 30, 1, NOW(), NOW()),
  ('报名名单', 'post:post', '', 3, @pid_post_post, '/api/admin/post/signup/list', 'GET', 40, 1, NOW(), NOW()),
  ('用户报名记录', 'post:post', '', 3, @pid_post_post, '/api/admin/post/signup/by-user', 'GET', 45, 1, NOW(), NOW()),
  ('帖子状态', 'post:post:status', '', 3, @pid_post_post, '/api/admin/post/update/status', 'POST', 50, 1, NOW(), NOW()),
  ('帖子批量状态', 'post:post:status', '', 3, @pid_post_post, '/api/admin/post/batch/status', 'POST', 60, 1, NOW(), NOW()),
  ('广告列表', 'ad:campaign', '', 3, @pid_ad_campaign, '/api/admin/ad/campaign/list', 'GET', 10, 1, NOW(), NOW()),
  ('广告详情', 'ad:campaign', '', 3, @pid_ad_campaign, '/api/admin/ad/campaign/detail', 'GET', 20, 1, NOW(), NOW()),
  ('广告保存', 'ad:campaign:update', '', 3, @pid_ad_campaign, '/api/admin/ad/campaign/save', 'POST', 30, 1, NOW(), NOW()),
  ('广告状态', 'ad:campaign:status', '', 3, @pid_ad_campaign, '/api/admin/ad/campaign/update/status', 'POST', 40, 1, NOW(), NOW()),
  ('广告批量暂停', 'ad:campaign:pause', '', 3, @pid_ad_campaign, '/api/admin/ad/campaign/batch/pause', 'POST', 50, 1, NOW(), NOW()),
  ('广告复制', 'ad:campaign:copy', '', 3, @pid_ad_campaign, '/api/admin/ad/campaign/copy', 'POST', 60, 1, NOW(), NOW()),
  ('广告日报', 'ad:campaign:stats', '', 3, @pid_ad_campaign, '/api/admin/ad/campaign/daily-stats', 'GET', 70, 1, NOW(), NOW()),
  ('广告商家列表', 'ad:campaign', '', 3, @pid_ad_campaign, '/api/admin/ad/campaign/advertisers', 'GET', 80, 1, NOW(), NOW()),
  ('广告审核', 'ad:campaign:audit', '', 3, @pid_ad_campaign, '/api/admin/ad/campaign/audit', 'POST', 90, 1, NOW(), NOW()),
  ('投放设置读取', 'ad:setting', '', 3, @pid_ad_setting, '/api/admin/ad/setting', 'GET', 10, 1, NOW(), NOW()),
  ('投放设置保存', 'ad:setting:update', '', 3, @pid_ad_setting, '/api/admin/ad/setting', 'POST', 20, 1, NOW(), NOW());

INSERT INTO `employee` (`mobile`, `name`, `employee_no`, `department_id`, `status`, `created_at`, `updated_at`) VALUES
  ('17000000000', '超级管理员', 'SA001', 0, 1, NOW(), NOW());

SET @super_admin_role_id = (SELECT `id` FROM `role` WHERE `code` = 'super_admin' LIMIT 1);
SET @super_admin_employee_id = (SELECT `id` FROM `employee` WHERE `mobile` = '17000000000' LIMIT 1);

INSERT INTO `employee_role` (`employee_id`, `role_id`, `created_at`, `updated_at`) VALUES
  (@super_admin_employee_id, @super_admin_role_id, NOW(), NOW());

INSERT INTO `app_update_policy` (`platform`, `min_supported_version`, `title`, `message`, `created_at`, `updated_at`) VALUES
  ('android', '1.0.0', '发现新版本', '请更新到最新版本以继续使用', NOW(), NOW()),
  ('ios', '1.0.0', '发现新版本', '请更新到最新版本以继续使用', NOW(), NOW()),
  ('h5', '1.0.0', '发现新版本', '请刷新页面以使用最新版本', NOW(), NOW());

INSERT INTO `im_support_config` (`brand_name`, `welcome_text`, `offline_text`, `enabled`, `quick_replies`, `created_at`, `updated_at`) VALUES
  ('官方客服', '您好，请描述您的问题，我们会尽快回复', '当前客服繁忙或不在线，留言后我们会尽快回复', 1, '[]', NOW(), NOW());

SET FOREIGN_KEY_CHECKS = 1;
