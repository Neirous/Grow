const Settings = {
    async load() {
        try {
            const settings = await App.api('/api/settings');
            this.render(settings);
        } catch (err) {
            App.toast('加载设置失败: ' + err.message, 'error');
        }
    },

    render(settings) {
        const container = document.getElementById('tab-settings');
        if (!container) return;

        container.innerHTML = `
            <div class="section-header"><h2>飞书提醒</h2></div>
            <div class="settings-card">
                <div class="form-group">
                    <label>飞书 Webhook URL</label>
                    <input type="text" id="s-feishu-url" class="input"
                        value="${this.esc(settings.feishu_webhook_url || '')}"
                        placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/xxx">
                </div>
                <div class="form-row">
                    <div class="form-group">
                        <label>提醒时间</label>
                        <input type="time" id="s-reminder-time" class="input"
                            value="${settings.reminder_time || '20:00'}">
                    </div>
                    <div class="form-group">
                        <label>启用提醒</label>
                        <select id="s-reminder-enabled" class="input">
                            <option value="false" ${settings.reminder_enabled !== 'true' ? 'selected' : ''}>关闭</option>
                            <option value="true" ${settings.reminder_enabled === 'true' ? 'selected' : ''}>开启</option>
                        </select>
                    </div>
                </div>
                ${settings.reminder_last_sent ? `<p style="font-size:12px;color:var(--text-muted)">上次发送: ${settings.reminder_last_sent}</p>` : ''}
            </div>

            <div class="section-header" style="margin-top:24px"><h2>Notion 同步</h2></div>
            <div class="settings-card">
                <div class="form-group">
                    <label>Notion API Key (Integration Token)</label>
                    <input type="password" id="s-notion-key" class="input"
                        value="${this.esc(settings.notion_api_key || '')}"
                        placeholder="secret_xxx">
                </div>
                <div class="form-group">
                    <label>Notion Database ID</label>
                    <input type="text" id="s-notion-db" class="input"
                        value="${this.esc(settings.notion_database_id || '')}"
                        placeholder="从 Notion 数据库 URL 中获取">
                </div>
                <div class="form-row">
                    <div class="form-group">
                        <label>自动同步</label>
                        <select id="s-notion-enabled" class="input">
                            <option value="false" ${settings.notion_sync_enabled !== 'true' ? 'selected' : ''}>关闭</option>
                            <option value="true" ${settings.notion_sync_enabled === 'true' ? 'selected' : ''}>每天 23:00 自动同步</option>
                        </select>
                    </div>
                    <div class="form-group" style="display:flex;align-items:flex-end">
                        <button class="btn btn-primary" onclick="Settings.syncNow()">立即同步</button>
                    </div>
                </div>
                ${settings.notion_last_sync ? `<p style="font-size:12px;color:var(--text-muted)">上次同步: ${settings.notion_last_sync}</p>` : ''}
                <div id="notion-sync-status" style="margin-top:8px"></div>
            </div>

            <div class="form-actions" style="margin-top:24px">
                <button class="btn btn-primary" id="settings-save">保存设置</button>
            </div>
        `;

        document.getElementById('settings-save').onclick = async () => {
            const data = {
                feishu_webhook_url: document.getElementById('s-feishu-url').value,
                reminder_time: document.getElementById('s-reminder-time').value,
                reminder_enabled: document.getElementById('s-reminder-enabled').value,
                notion_api_key: document.getElementById('s-notion-key').value,
                notion_database_id: document.getElementById('s-notion-db').value,
                notion_sync_enabled: document.getElementById('s-notion-enabled').value,
            };
            try {
                await App.api('/api/settings', { method: 'PUT', body: JSON.stringify(data) });
                App.toast('设置已保存');
            } catch (err) {
                App.toast('保存失败: ' + err.message, 'error');
            }
        };
    },

    async syncNow() {
        const statusEl = document.getElementById('notion-sync-status');
        statusEl.innerHTML = '<span style="color:var(--primary)">同步中...</span>';
        try {
            const result = await App.api('/api/export/notion', { method: 'POST' });
            statusEl.innerHTML = `<span style="color:var(--success)">同步完成！更新 ${result.updated}，新建 ${result.created}，共 ${result.total} 项能力</span>`;
        } catch (err) {
            statusEl.innerHTML = `<span style="color:var(--danger)">同步失败: ${err.message}</span>`;
        }
    },

    esc(s) {
        const div = document.createElement('div');
        div.textContent = s || '';
        return div.innerHTML;
    }
};
