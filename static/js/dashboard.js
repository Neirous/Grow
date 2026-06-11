const Dashboard = {
    async load() {
        try {
            const data = await App.api('/api/dashboard');
            this.renderAbilities(data.abilities);
            this.renderQuickComplete(data.activities);
            this.renderGoals();
            this.loadStreaks();
            this.loadRecentLogs();
        } catch (err) {
            App.toast('加载仪表盘失败: ' + err.message, 'error');
        }
    },

    async renderGoals() {
        const container = document.getElementById('goal-cards');
        try {
            const goals = await App.api('/api/goals');
            if (!goals || goals.length === 0) {
                container.innerHTML = '<div class="empty-state" style="font-size:13px;padding:12px"><p>还没有目标，点击「设定目标」创建一个</p></div>';
                return;
            }

            container.innerHTML = goals.map(g => {
                const progress = Math.min(100, Math.max(0, g.progress || 0));
                const remaining = g.target_value - g.current_value;
                const deadlineStr = g.deadline
                    ? `截止 ${new Date(g.deadline).toLocaleDateString('zh-CN')}`
                    : '无截止日期';

                return `
                    <div class="goal-card">
                        <div class="goal-card-header">
                            <span class="goal-card-name">${this.esc(g.ability_name)}</span>
                            <span class="goal-card-deadline">${deadlineStr}</span>
                        </div>
                        <div class="goal-card-bar">
                            <div class="goal-progress-fill" style="width:${progress}%"></div>
                        </div>
                        <div class="goal-card-stats">
                            <span>目标 ${App.formatNumber(g.target_value)}</span>
                            <span>当前 ${App.formatNumber(g.current_value)}</span>
                            <span>${progress.toFixed(0)}%</span>
                            ${remaining > 0 ? `<span class="goal-remaining">还剩 ${App.formatNumber(remaining)}</span>` : '<span class="goal-done">已达成 ✅</span>'}
                        </div>
                    </div>
                `;
            }).join('');
        } catch (err) {
            container.innerHTML = '<div class="empty-state" style="font-size:13px;padding:12px"><p>加载目标失败</p></div>';
        }
    },

    async loadStreaks() {
        const container = document.getElementById('streak-heatmap');
        try {
            const data = await App.api('/api/streaks');
            Heatmap.render('streak-heatmap', data);
        } catch (err) {
            container.innerHTML = '';
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
                    <div class="ability-card-name">${this.esc(a.name)}</div>
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
            activities.map(a => `<option value="${a.id}">${this.esc(a.name)}</option>`).join('');

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
                this.load();
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
                        <span class="log-entry-name">${this.esc(log.activity_name)}</span>
                        <span class="log-entry-time">${App.timeAgo(log.completed_at)}</span>
                    </div>
                    <div class="log-entry-changes">
                        ${(log.snapshots || []).map(s => `
                            <span class="log-change log-change-up">
                                ${this.esc(s.ability_name)}:
                                ${App.formatNumber(s.old_value)} → ${App.formatNumber(s.new_value)}
                                (+${((s.new_value - s.old_value) / s.old_value * 100).toFixed(2)}%)
                            </span>
                        `).join('')}
                    </div>
                    ${log.note ? `<div class="log-entry-note">📝 ${this.esc(log.note)}</div>` : ''}
                </div>
            `).join('');
        } catch (err) {
            container.innerHTML = '<div class="empty-state"><p>加载日志失败</p></div>';
        }
    },

    esc(s) {
        const div = document.createElement('div');
        div.textContent = s || '';
        return div.innerHTML;
    }
};

const Goals = {
    async showForm(id) {
        let abilities = [];
        try { abilities = await App.api('/api/abilities'); } catch (e) {}

        Modal.show('设定目标', `
            <div class="form-group">
                <label>能力 *</label>
                <select id="g-ability" class="input">
                    <option value="">-- 选择能力 --</option>
                    ${abilities.map(a => `<option value="${a.id}">${Goals.esc(a.name)} (当前: ${App.formatNumber(a.effective_value || a.current_value)})</option>`).join('')}
                </select>
            </div>
            <div class="form-group">
                <label>目标值 *</label>
                <input type="number" id="g-target" class="input" step="1" placeholder="如 2000">
            </div>
            <div class="form-group">
                <label>截止日期（可选）</label>
                <input type="date" id="g-deadline" class="input">
            </div>
            <div class="form-actions">
                <button class="btn btn-secondary" onclick="Modal.hide()">取消</button>
                <button class="btn btn-primary" id="g-save">保存</button>
            </div>
        `);

        document.getElementById('g-save').onclick = async () => {
            const data = {
                ability_id: parseInt(document.getElementById('g-ability').value),
                target_value: parseFloat(document.getElementById('g-target').value),
                deadline: document.getElementById('g-deadline').value
            };
            if (!data.ability_id || !data.target_value) {
                App.toast('请填写能力和目标值', 'warning');
                return;
            }
            try {
                await App.api('/api/goals', { method: 'POST', body: JSON.stringify(data) });
                Modal.hide();
                App.toast('目标已创建');
                Dashboard.renderGoals();
            } catch (err) {
                App.toast('创建失败: ' + err.message, 'error');
            }
        };
    },

    async deleteGoal(id) {
        if (!confirm('删除这个目标？')) return;
        try {
            await App.api(`/api/goals/${id}`, { method: 'DELETE' });
            App.toast('目标已删除');
            Dashboard.renderGoals();
        } catch (err) {
            App.toast('删除失败: ' + err.message, 'error');
        }
    },

    esc(s) {
        const div = document.createElement('div');
        div.textContent = s || '';
        return div.innerHTML;
    }
};
