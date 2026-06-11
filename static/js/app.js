const App = {
    currentTab: 'dashboard',

    init() {
        this.cacheDOM();
        this.bindNavigation();
        this.switchTab('dashboard');
    },

    cacheDOM() {
        this.tabContents = document.querySelectorAll('.tab-content');
        this.navLinks = document.querySelectorAll('[data-tab]');
    },

    bindNavigation() {
        this.navLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const tab = e.currentTarget.dataset.tab;
                this.switchTab(tab);
            });
        });
    },

    switchTab(tab) {
        this.tabContents.forEach(el => el.hidden = true);
        const target = document.getElementById(`tab-${tab}`);
        if (target) target.hidden = false;
        this.navLinks.forEach(el => el.classList.remove('active'));
        const navEl = document.querySelector(`[data-tab="${tab}"]`);
        if (navEl) navEl.classList.add('active');
        this.currentTab = tab;
        this.loadTabContent(tab);
    },

    loadTabContent(tab) {
        switch(tab) {
            case 'dashboard': Dashboard.load(); break;
            case 'abilities': Abilities.load(); break;
            case 'activities': Activities.load(); break;
            case 'logs': Logs.load(); break;
        }
    },

    async api(path, options = {}) {
        try {
            const resp = await fetch(path, {
                headers: { 'Content-Type': 'application/json', ...options.headers },
                ...options
            });
            if (!resp.ok) {
                const err = await resp.json().catch(() => ({error: resp.statusText}));
                throw new Error(err.error || `HTTP ${resp.status}`);
            }
            return resp.json();
        } catch (err) {
            if (err.name === 'TypeError' && err.message === 'Failed to fetch') {
                throw new Error('无法连接到服务器');
            }
            throw err;
        }
    },

    toast(message, type = 'success') {
        const container = document.getElementById('toast-container');
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;
        container.appendChild(toast);
        setTimeout(() => {
            toast.style.opacity = '0';
            toast.style.transition = 'opacity 0.3s';
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    },

    formatNumber(n) {
        if (n === undefined || n === null) return '0';
        return Number(n).toFixed(2);
    },

    timeAgo(dateStr) {
        if (!dateStr) return '从未';
        const now = new Date();
        const date = new Date(dateStr + (dateStr.includes('T') ? '' : 'Z'));
        if (isNaN(date.getTime())) return dateStr;
        const diffMs = now - date;
        const diffMin = Math.floor(diffMs / 60000);
        const diffHours = Math.floor(diffMs / 3600000);
        const diffDays = Math.floor(diffMs / 86400000);

        if (diffMin < 1) return '刚刚';
        if (diffMin < 60) return `${diffMin}分钟前`;
        if (diffHours < 24) return `${diffHours}小时前`;
        if (diffDays < 30) return `${diffDays}天前`;
        return date.toLocaleDateString('zh-CN');
    },

    formatDate(dateStr) {
        if (!dateStr) return '';
        const d = new Date(dateStr + (dateStr.includes('T') ? '' : 'Z'));
        if (isNaN(d.getTime())) return dateStr;
        return d.toLocaleString('zh-CN');
    },

    fetchSelectOptions(selectId, items, valueKey, labelKey, placeholder) {
        const select = document.getElementById(selectId);
        if (!select) return;
        select.innerHTML = `<option value="">${placeholder || '-- 请选择 --'}</option>`;
        items.forEach(item => {
            const opt = document.createElement('option');
            opt.value = item[valueKey];
            opt.textContent = item[labelKey];
            select.appendChild(opt);
        });
    }
};

const Modal = {
    show(title, bodyHTML) {
        document.getElementById('modal-title').textContent = title;
        document.getElementById('modal-body').innerHTML = bodyHTML;
        document.getElementById('modal-overlay').hidden = false;
    },

    hide() {
        document.getElementById('modal-overlay').hidden = true;
        document.getElementById('modal-body').innerHTML = '';
    }
};

document.getElementById('modal-overlay').addEventListener('click', (e) => {
    if (e.target === e.currentTarget) Modal.hide();
});

document.addEventListener('DOMContentLoaded', () => App.init());
