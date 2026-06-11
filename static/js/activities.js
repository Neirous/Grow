const Activities = {
    abilitiesCache: [],

    async load() {
        try {
            const activities = await App.api('/api/activities');
            const abilities = await App.api('/api/abilities');
            this.abilitiesCache = abilities;
            this.renderList(activities);
        } catch (err) {
            App.toast('加载活动列表失败: ' + err.message, 'error');
        }
    },

    renderList(activities) {
        const container = document.getElementById('activities-list');

        if (!activities || activities.length === 0) {
            container.innerHTML = '<div class="empty-state"><div class="empty-state-icon">🏃</div><p>还没有活动，点击"+ 新增活动"创建第一个</p></div>';
            return;
        }

        container.innerHTML = activities.map(a => `
            <div class="activity-row">
                <div class="activity-row-info">
                    <h4>${this.escapeHtml(a.name)}</h4>
                    <div>
                        ${(a.effects || []).map(e =>
                            `<span class="effect-badge">${this.escapeHtml(e.ability_name)} +${App.formatNumber(e.boost_percentage)}%</span>`
                        ).join('')}
                        ${(!a.effects || a.effects.length === 0) ? '<span style="color:var(--text-muted);font-size:12px">无效果</span>' : ''}
                    </div>
                    ${a.description ? `<p style="font-size:12px;color:var(--text-secondary);margin-top:4px">${this.escapeHtml(a.description)}</p>` : ''}
                </div>
                <div class="row-actions">
                    <button class="btn btn-success btn-sm" onclick="Activities.completeActivity(${a.id})">完成</button>
                    <button class="btn btn-secondary btn-sm" onclick="Activities.showForm(${a.id})">编辑</button>
                    <button class="btn btn-danger btn-sm" onclick="Activities.deleteActivity(${a.id})">删除</button>
                </div>
            </div>
        `).join('');
    },

    async completeActivity(id) {
        const note = prompt('备注（可选）:');
        if (note === null) return; // cancelled
        try {
            const result = await App.api(`/api/activities/${id}/complete`, {
                method: 'POST',
                body: JSON.stringify({ note: note || '' })
            });
            App.toast('活动完成！' + result.snapshots.map(s =>
                `${s.ability_name}: ${App.formatNumber(s.old_value)} → ${App.formatNumber(s.new_value)}`
            ).join(', '));
            this.load();
        } catch (err) {
            App.toast('完成失败: ' + err.message, 'error');
        }
    },

    async showForm(id) {
        let activity = { name: '', description: '', effects: [] };
        let isEdit = false;

        if (id) {
            isEdit = true;
            try {
                activity = await App.api(`/api/activities/${id}`);
            } catch (err) {
                App.toast('加载活动失败', 'error');
                return;
            }
        }

        // Load abilities if not cached
        if (this.abilitiesCache.length === 0) {
            try {
                this.abilitiesCache = await App.api('/api/abilities');
            } catch (err) { /* ignore */ }
        }

        const abilitiesOptions = this.abilitiesCache.map(a =>
            `<option value="${a.id}">${this.escapeHtml(a.name)}</option>`
        ).join('');

        Modal.show(isEdit ? '编辑活动' : '新建活动', `
            <div class="form-group">
                <label>名称 *</label>
                <input type="text" id="form-name" value="${this.escapeHtml(activity.name)}" class="input">
            </div>
            <div class="form-group">
                <label>描述</label>
                <textarea id="form-desc" class="input">${this.escapeHtml(activity.description || '')}</textarea>
            </div>
            <div class="form-group">
                <label>效果列表（能力 + 提升百分比）</label>
                <div id="effect-manager" class="effect-manager"></div>
                <button class="btn btn-secondary btn-sm" id="add-effect" style="margin-top:8px">+ 添加效果</button>
            </div>
            <div class="form-actions">
                <button class="btn btn-secondary" onclick="Modal.hide()">取消</button>
                <button class="btn btn-primary" id="form-save">保存</button>
            </div>
        `);

        // Render existing effects
        const effectContainer = document.getElementById('effect-manager');
        const renderEffectRow = (effect) => {
            const row = document.createElement('div');
            row.className = 'effect-row';
            row.innerHTML = `
                <select class="input effect-ability">
                    <option value="">-- 选择能力 --</option>
                    ${abilitiesOptions}
                </select>
                <input type="number" class="input effect-boost" value="${effect ? App.formatNumber(effect.boost_percentage) : '1.0'}" step="0.1" placeholder="提升%">
                <button class="btn btn-danger btn-xs remove-effect">✕</button>
            `;
            if (effect) {
                row.querySelector('.effect-ability').value = effect.ability_id;
            }
            row.querySelector('.remove-effect').onclick = () => row.remove();
            return row;
        };

        (activity.effects || []).forEach(e => {
            effectContainer.appendChild(renderEffectRow(e));
        });

        document.getElementById('add-effect').onclick = () => {
            effectContainer.appendChild(renderEffectRow(null));
        };

        document.getElementById('form-save').onclick = async () => {
            const name = document.getElementById('form-name').value;
            if (!name) { App.toast('请输入活动名称', 'warning'); return; }

            const effects = [];
            effectContainer.querySelectorAll('.effect-row').forEach(row => {
                const abilityId = parseInt(row.querySelector('.effect-ability').value);
                const boost = parseFloat(row.querySelector('.effect-boost').value);
                if (abilityId && boost) {
                    effects.push({ ability_id: abilityId, boost_percentage: boost });
                }
            });

            const data = {
                name,
                description: document.getElementById('form-desc').value,
                effects
            };

            try {
                if (isEdit) {
                    await App.api(`/api/activities/${id}`, { method: 'PUT', body: JSON.stringify(data) });
                } else {
                    await App.api('/api/activities', { method: 'POST', body: JSON.stringify(data) });
                }
                Modal.hide();
                App.toast(isEdit ? '活动已更新' : '活动已创建');
                this.load();
            } catch (err) {
                App.toast('保存失败: ' + err.message, 'error');
            }
        };
    },

    async deleteActivity(id) {
        if (!confirm('确定要删除这个活动吗？')) return;
        try {
            await App.api(`/api/activities/${id}`, { method: 'DELETE' });
            App.toast('活动已删除');
            this.load();
        } catch (err) {
            App.toast('删除失败: ' + err.message, 'error');
        }
    },

    escapeHtml(str) {
        const div = document.createElement('div');
        div.textContent = str || '';
        return div.innerHTML;
    }
};
