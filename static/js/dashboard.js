const Dashboard = {
    async load() {
        try {
            const data = await App.api('/api/dashboard');
            this.renderAbilities(data.abilities);
            this.renderQuickComplete(data.activities);
            this.loadRecentLogs();
        } catch (err) {
            App.toast('加载仪表盘失败: ' + err.message, 'error');
        }
    },

    renderAbilities(abilities) {
        const container = document.getElementById('ability-cards');
        if (!abilities || abilities.length === 0) {
            container.innerHTML = '<div class="empty-state"><div class="empty-state-icon">📊</div><p>还没有能力数据，去"能力管理"创建一个吧</p></div>';
            return;
        }

        container.innerHTML = abilities.map(a => {
            const decayNote = a.days_since_last_activity > 0
                ? `<span class="decay-warn">衰减中 (${a.days_since_last_activity.toFixed(1)}天未练)</span>`
                : '<span>✔ 今日已练</span>';
            const effectiveNote = a.days_since_last_activity > 0 && a.effective_value !== a.current_value
                ? `<div class="ability-card-effective">有效值: ${App.formatNumber(a.effective_value)} (衰减后)</div>`
                : '';

            return `
                <div class="ability-card" onclick="Abilities.showDetail(${a.id})">
                    <div class="ability-card-name">${this.escapeHtml(a.name)}</div>
                    <div class="ability-card-value">${App.formatNumber(a.effective_value || a.current_value)}</div>
                    ${effectiveNote}
                    <div class="ability-card-meta">
                        <span>基础: ${App.formatNumber(a.base_value)}</span>
                        ${decayNote}
                    </div>
                </div>
            `;
        }).join('');
    },

    renderQuickComplete(activities) {
        const select = document.getElementById('quick-activity-select');
        select.innerHTML = '<option value="">-- 选择活动 --</option>' +
            activities.map(a => `<option value="${a.id}">${this.escapeHtml(a.name)}</option>`).join('');

        document.getElementById('quick-complete-btn').onclick = async () => {
            const activityId = select.value;
            if (!activityId) {
                App.toast('请先选择一个活动', 'warning');
                return;
            }
            const note = document.getElementById('quick-note').value;
            try {
                const result = await App.api(`/api/activities/${activityId}/complete`, {
                    method: 'POST',
                    body: JSON.stringify({ note })
                });
                document.getElementById('quick-note').value = '';
                App.toast('活动完成！' + result.snapshots.map(s =>
                    `${s.ability_name}: ${App.formatNumber(s.old_value)} → ${App.formatNumber(s.new_value)}`
                ).join(', '));
                this.load(); // refresh dashboard
            } catch (err) {
                App.toast('完成失败: ' + err.message, 'error');
            }
        };
    },

    async loadRecentLogs() {
        const container = document.getElementById('recent-logs');
        try {
            const logs = await App.api('/api/logs?limit=10');
            if (!logs || logs.length === 0) {
                container.innerHTML = '<div class="empty-state"><p>还没有活动记录</p></div>';
                return;
            }

            container.innerHTML = logs.map(log => `
                <div class="log-entry">
                    <div class="log-entry-header">
                        <span class="log-entry-name">${this.escapeHtml(log.activity_name)}</span>
                        <span class="log-entry-time">${App.timeAgo(log.completed_at)}</span>
                    </div>
                    <div class="log-entry-changes">
                        ${(log.snapshots || []).map(s => `
                            <span class="log-change log-change-up">
                                ${this.escapeHtml(s.ability_name)}:
                                ${App.formatNumber(s.old_value)} → ${App.formatNumber(s.new_value)}
                                (+${((s.new_value - s.old_value) / s.old_value * 100).toFixed(2)}%)
                            </span>
                        `).join('')}
                    </div>
                    ${log.note ? `<div class="log-entry-note">📝 ${this.escapeHtml(log.note)}</div>` : ''}
                </div>
            `).join('');
        } catch (err) {
            container.innerHTML = '<div class="empty-state"><p>加载日志失败</p></div>';
        }
    },

    escapeHtml(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }
};
