(function() {
    'use strict';

    const HomeApp = {
        ws: null,
        reconnectTimer: null,

        utils: {
            escapeHtml: function(text) {
                if (!text) return '';
                const map = {
                    '&': '&amp;',
                    '<': '&lt;',
                    '>': '&gt;',
                    '"': '&quot;',
                    "'": '&#039'
                };
                return text.replace(/[&<>'"]/g, function(m) { return map[m]; });
            },

            getTypeIcon: function(type) {
                const icons = {
                    'http': '<i class="fas fa-globe"></i>',
                    'https': '<i class="fas fa-globe"></i>',
                    'tcp': '<i class="fas fa-network-wired"></i>',
                    'udp': '<i class="fas fa-broadcast-tower"></i>',
                    'dns': '<i class="fas fa-server"></i>',
                    'ping': '<i class="fas fa-wifi"></i>'
                };
                return icons[type] || '<i class="fas fa-question"></i>';
            },

            padZero: function(num) {
                return String(num).padStart(2, '0');
            },

            formatDate: function(dateStr) {
                const date = new Date(dateStr);
                return date.toLocaleDateString('zh-CN', {
                    year: 'numeric',
                    month: '2-digit',
                    day: '2-digit',
                    hour: '2-digit',
                    minute: '2-digit',
                    second: '2-digit'
                });
            },

            formatSize: function(bytes) {
                if (bytes >= 1024) {
                    return (bytes / 1024).toFixed(2) + 'kb';
                }
                return bytes + 'b';
            },

            formatSpeed: function(speed) {
                if (speed === undefined || speed === null) {
                    return '';
                }
                return speed.toFixed(2) + 'ms';
            },

            getUrlParam: function(name) {
                return new URLSearchParams(window.location.search).get(name);
            },

            getWsUrl: function(path, params) {
                const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
                let url = protocol + '//' + window.location.host + path;
                if (params) {
                    url += '?' + params;
                }
                return url;
            }
        },

        connect: function(wsUrl, handlers) {
            const self = this;
            try {
                this.ws = new WebSocket(wsUrl);

                this.ws.onopen = handlers.onopen || function() {};
                this.ws.onmessage = handlers.onmessage || function() {};
                this.ws.onclose = handlers.onclose || function() {};
                this.ws.onerror = handlers.onerror || function() {};
            } catch (e) {
                console.error('WebSocket connection failed:', e);
                if (this.reconnectTimer) {
                    clearTimeout(this.reconnectTimer);
                }
                this.reconnectTimer = setTimeout(function() {
                    self.connect(wsUrl, handlers);
                }, 5000);
            }
        },

        reconnect: function(wsUrl, handlers) {
            if (this.reconnectTimer) {
                clearTimeout(this.reconnectTimer);
            }
            this.reconnectTimer = setTimeout(function() {
                this.connect(wsUrl, handlers);
            }.bind(this), 5000);
        },

        // 渲染时间网格
        renderTimeGrid: function(hourLogs) {
            let html = '<div class="time-grid">';

            if (hourLogs && hourLogs.length > 0) {
                for (let i = 0; i < hourLogs.length; i++) {
                    const log = hourLogs[i];
                    const status = log.is_valid ? 'up' : 'down';
                    const error = log.error_msg || '无法访问';
                    const timeStr = HomeApp.utils.padZero(log.hour) + ':' + HomeApp.utils.padZero(log.minute || 0);
                    let cellHtml = '<div class="time-cell ' + status + '" data-status="' + status + '" data-time="' + timeStr + '" data-error="' + HomeApp.utils.escapeHtml(error) + '"';
                    if (log.speed) {
                        cellHtml += ' data-speed="' + log.speed + '"';
                    }
                    if (log.size !== undefined && log.size !== null) {
                        let sizeStr;
                        if (log.size >= 1024) {
                            sizeStr = (log.size / 1024).toFixed(2) + 'kb';
                        } else {
                            sizeStr = log.size + 'b';
                        }
                        cellHtml += ' data-size="' + sizeStr + '"';
                    }
                    cellHtml += '></div>';
                    html += cellHtml;
                }
            } else {
                // 生成模拟数据
                for (let i = 0; i < 468; i++) { // 6行 * 78列
                    const isUp = Math.random() > 0.2;
                    const status = isUp ? 'up' : 'down';
                    const hour = HomeApp.utils.padZero(Math.floor(i / 78));
                    const minute = HomeApp.utils.padZero(Math.floor((i % 78) * (60 / 78)));
                    const timeStr = hour + ':' + minute;
                    const speed = (Math.random() * 1000).toFixed(2) + 'ms';
                    const size = (Math.random() * 200).toFixed(2) + 'kb';
                    
                    html += '<div class="time-cell ' + status + '" ' +
                        'data-status="' + status + '" ' +
                        'data-time="' + timeStr + '" ' +
                        'data-speed="' + speed + '" ' +
                        'data-size="' + size + '"></div>';
                }
            }

            html += '</div>';
            return html;
        },

        // 渲染图例
        renderLegend: function() {
            return '<div class="legend">' +
                '<div class="legend-item"><div class="legend-color" style="background-color: #27ae60;"></div><span>正常</span></div>' +
                '<div class="legend-item"><div class="legend-color" style="background-color: #e74c3c;"></div><span>故障</span></div>' +
                '<div class="legend-item"><div class="legend-color" style="background-color: #f0f0f0;"></div><span>未知</span></div>' +
                '</div>';
        },

        // 渲染状态卡片
        renderStatusCard: function(monitor, showJump) {
            const statusClass = monitor.is_valid ? 'up' : 'down';
            const statusText = monitor.is_valid ? '正常' : '无法访问';
            const statusIcon = monitor.is_valid ? 'fa-check' : 'fa-times';

            let html = '<div class="status-card" data-category="tab-' + monitor.gid + '" data-id="' + monitor.id + '">';
            html += '<div class="status-header">';
            html += '<div class="service-name">' + HomeApp.utils.getTypeIcon(monitor.type) + ' ' + HomeApp.utils.escapeHtml(monitor.name) + '</div>';
            html += '<div class="status-indicator ' + statusClass + '">';
            html += '<i class="fas ' + statusIcon + '"></i>';
            html += '<span>' + statusText + '</span>';
            if (monitor.latency) {
                html += '<span class="latency">' + monitor.latency + '</span>';
            }
            html += '</div>';
            html += '</div>';

            // 渲染时间网格
            html += this.renderTimeGrid(monitor.list);

            // 渲染图例
            html += this.renderLegend();

            // 渲染页脚统计
            html += '<div class="status-footer">';
            html += '<div class="status-stats">类型: ' + HomeApp.utils.escapeHtml(monitor.type) + '</div>';
            html += '<div class="status-stats">日志: ' + (monitor.list ? monitor.list.length : 0) + ' 条</div>';
            if (monitor.is_valid) {
                if (monitor.speed) {
                    html += '<div class="status-stats">耗时: ' + monitor.speed + 'ms</div>';
                }
                if (monitor.size !== undefined && monitor.size !== null) {
                    let sizeStr;
                    if (monitor.size >= 1024) {
                        sizeStr = (monitor.size / 1024).toFixed(2) + 'kb';
                    } else {
                        sizeStr = monitor.size + 'b';
                    }
                    html += '<div class="status-stats">大小: ' + sizeStr + '</div>';
                }
            }
            // 总是显示可用率，即使为0
            html += '<div class="status-stats">可用率: ' + (monitor.up_rate || 0).toFixed(1) + '%</div>';
            if (showJump) {
                html += '<div class="status-jump"><a href="/monitor?id=' + monitor.id + '" style="color: #3498db; text-decoration: none;"><i class="fas fa-arrow-right"></i></a></div>';
            }
            html += '</div>';
            // 为状态指示器添加错误信息属性，用于 tooltip
            html += '<div class="error-tooltip" style="display: none;" data-error="' + HomeApp.utils.escapeHtml(monitor.error_msg || '') + '"></div>';

            html += '</div>';

            return html;
        },

        // 清理旧的提示元素
        clearTips: function() {
            const oldTips = document.querySelectorAll('.time-cell-tip, .status-indicator-tip');
            oldTips.forEach(tip => tip.remove());
        },

        // 绑定标签事件
        bindTabEvents: function() {
            const tabs = document.querySelectorAll('.tab');
            const statusCards = document.querySelectorAll('.status-card');

            tabs.forEach(tab => {
                tab.addEventListener('click', function() {
                    tabs.forEach(t => t.classList.remove('active'));
                    this.classList.add('active');

                    const tabName = this.dataset.tab;

                    statusCards.forEach(card => {
                        if (tabName === 'all' || card.dataset.category === tabName) {
                            card.style.display = 'block';
                        } else {
                            card.style.display = 'none';
                        }
                    });
                });
            });
        },

        // 渲染标签栏
        renderTabs: function(groups) {
            const tabsContainer = document.getElementById('status-tabs');
            if (!tabsContainer) return;

            let html = '<div class="tab active" data-tab="all">All</div>';

            if (groups && groups.length > 0) {
                groups.forEach(function(group) {
                    html += '<div class="tab" data-tab="tab-' + group.id + '">' + group.name + '</div>';
                });
            }

            tabsContainer.innerHTML = html;
            this.bindTabEvents();
        },

        // 渲染 index 页面
        renderIndex: {
            data: null,
            groups: null,

            init: function() {
                const self = this;
                const wsUrl = HomeApp.utils.getWsUrl('/ws/status');

                const handlers = {
                    onopen: function() {
                        console.log('Index WebSocket connected');
                    },
                    onmessage: function(event) {
                        try {
                            const data = JSON.parse(event.data);
                            if (data.type === 'monitor_status') {
                                const activeTab = document.querySelector('.tab.active');
                                const activeTabName = activeTab ? activeTab.dataset.tab : 'all';

                                self.data = data.data;
                                self.groups = data.groups || [];
                                self.renderStatusCards();
                                self.renderTabs();

                                if (activeTabName) {
                                    const tabToActivate = document.querySelector('.tab[data-tab="' + activeTabName + '"]') || document.querySelector('.tab[data-tab="all"]');
                                    if (tabToActivate) {
                                        tabToActivate.click();
                                    }
                                }
                            }
                        } catch (e) {
                            console.error('Failed to parse message:', e);
                        }
                    },
                    onclose: function() {
                        console.log('Index WebSocket disconnected, reconnecting...');
                        HomeApp.reconnect(wsUrl, handlers);
                    },
                    onerror: function(error) {
                        console.error('Index WebSocket error:', error);
                    }
                };
                HomeApp.connect(wsUrl, handlers);
            },

            renderStatusCards: function() {
                const container = document.getElementById('status-container');
                if (!container) return;

                // 清理所有旧的提示元素
                HomeApp.clearTips();

                if (!this.data || this.data.length === 0) {
                    container.innerHTML = '<div class="loading">暂无监控数据</div>';
                    return;
                }

                let html = '';
                this.data.forEach(function(monitor) {
                    html += HomeApp.renderStatusCard(monitor, true);
                });

                container.innerHTML = html;

                // 重新绑定 tab 点击事件
                HomeApp.bindTabEvents();
            },

            renderTabs: function() {
                HomeApp.renderTabs(this.groups);
            }
        },

        // 渲染 groups 页面
        renderGroups: {
            data: null,

            init: function() {
                const self = this;
                const groupId = HomeApp.utils.getUrlParam('id');

                let wsUrl = HomeApp.utils.getWsUrl('/ws/groups');
                if (groupId) {
                    wsUrl += '?id=' + groupId;
                }

                const handlers = {
                    onopen: function() {
                        console.log('Groups WebSocket connected');
                    },
                    onmessage: function(event) {
                        try {
                            const data = JSON.parse(event.data);
                            if (data.type === 'group_monitors') {
                                self.data = data.data;
                                self.renderCards();
                            }
                        } catch (e) {
                            console.error('Failed to parse message:', e);
                        }
                    },
                    onclose: function() {
                        console.log('Groups WebSocket disconnected, reconnecting...');
                        HomeApp.reconnect(wsUrl, handlers);
                    },
                    onerror: function(error) {
                        console.error('Groups WebSocket error:', error);
                    }
                };
                HomeApp.connect(wsUrl, handlers);
            },

            renderCards: function() {
                const container = document.getElementById('group-container');
                if (!container) return;

                // 清理所有旧的提示元素
                HomeApp.clearTips();

                if (!this.data || this.data.length === 0) {
                    container.innerHTML = '<div class="loading">暂无分组数据</div>';
                    return;
                }

                const groupId = HomeApp.utils.getUrlParam('id');
                let groupsToDisplay = this.data;

                if (groupId) {
                    groupsToDisplay = this.data.filter(function(group) {
                        return group.id == groupId;
                    });

                    if (groupsToDisplay.length === 0) {
                        container.innerHTML = '<div class="loading">未找到指定的分组</div>';
                        return;
                    }
                }

                let html = '';
                groupsToDisplay.forEach(function(group) {
                    const monitors = group.monitors || [];
                    const upCount = monitors.filter(function(m) { return m.is_valid; }).length;
                    const totalCount = monitors.length;
                    const upRate = totalCount > 0 ? (upCount / totalCount * 100).toFixed(1) + '%' : '0.0%';

                    html += '<div class="group-section" data-group-id="' + group.id + '">';
                    html += '<h2>' + HomeApp.utils.escapeHtml(group.name) + '</h2>';
                    html += '<div>';
                    html += '正常: ' + upCount + '/' + totalCount + ' (' + upRate + ')';
                    html += '</div>';
                    if (monitors.length > 0) {
                        monitors.forEach(function(monitor) {
                            const statusClass = monitor.is_valid ? 'up' : 'down';
                            const statusText = monitor.is_valid ? '正常' : '无法访问';
                            const statusIcon = monitor.is_valid ? 'fa-check' : 'fa-times';

                            let cardHtml = '<div class="status-card" data-category="tab-' + monitor.gid + '" data-id="' + monitor.id + '">';
                            cardHtml += '<div class="status-header">';
                            cardHtml += '<div class="service-name">' + HomeApp.utils.getTypeIcon(monitor.type) + ' ' + HomeApp.utils.escapeHtml(monitor.name) + '</div>';
                            cardHtml += '<div class="status-indicator ' + statusClass + '">';
                            cardHtml += '<i class="fas ' + statusIcon + '"></i>';
                            cardHtml += '<span>' + statusText + '</span>';
                            if (monitor.latency) {
                                cardHtml += '<span class="latency">' + monitor.latency + '</span>';
                            }
                            cardHtml += '</div>';
                            cardHtml += '</div>';

                            // 渲染时间网格
                            cardHtml += HomeApp.renderTimeGrid(monitor.hour_logs);

                            // 渲染图例
                            cardHtml += HomeApp.renderLegend();

                            // 渲染页脚统计
                            cardHtml += '<div class="status-footer">';
                            cardHtml += '<div class="status-stats">类型: ' + HomeApp.utils.escapeHtml(monitor.type) + '</div>';
                            cardHtml += '<div class="status-stats">日志: ' + (monitor.hour_logs ? monitor.hour_logs.length : 0) + ' 条</div>';
                            if (monitor.is_valid) {
                                if (monitor.speed) {
                                    cardHtml += '<div class="status-stats">耗时: ' + monitor.speed + 'ms</div>';
                                }
                                if (monitor.size !== undefined && monitor.size !== null) {
                                    let sizeStr;
                                    if (monitor.size >= 1024) {
                                        sizeStr = (monitor.size / 1024).toFixed(2) + 'kb';
                                    } else {
                                        sizeStr = monitor.size + 'b';
                                    }
                                    cardHtml += '<div class="status-stats">大小: ' + sizeStr + '</div>';
                                }
                            }
                            // 总是显示可用率，即使为0
                            cardHtml += '<div class="status-stats">可用率: ' + (monitor.up_rate || 0).toFixed(1) + '%</div>';
                            cardHtml += '<div class="status-jump"><a href="/monitor?id=' + monitor.id + '" style="color: #3498db; text-decoration: none;"><i class="fas fa-arrow-right"></i></a></div>';
                            cardHtml += '</div>';
                            // 为状态指示器添加错误信息属性，用于 tooltip
                            cardHtml += '<div class="error-tooltip" style="display: none;" data-error="' + HomeApp.utils.escapeHtml(monitor.error_msg || '') + '"></div>';

                            cardHtml += '</div>';
                            html += cardHtml;
                        });
                    } else {
                        html += '<div style="text-align: center; padding: 40px 0; color: #999; font-size: 14px;">该分组下暂无监控点</div>';
                    }

                    html += '</div>';
                }.bind(this));

                container.innerHTML = html;
            }
        },

        // 渲染 monitor 页面
        renderMonitor: {
            data: null,
            historyDays: [],
            loadingComplete: false,
            initialized: false, // 标记是否已经初始化过

            init: function() {
                const self = this;
                const monitorId = HomeApp.utils.getUrlParam('id');

                if (!monitorId) {
                    document.getElementById('monitor-container').innerHTML = '<div class="loading">缺少监控点ID参数</div>';
                    return;
                }

                // 只在第一次初始化时清空历史数据
                if (!this.initialized) {
                    this.historyDays = [];
                    this.loadingComplete = false;
                    this.initialized = true;
                }

                // 根据是否已加载历史数据决定是否请求历史数据
                let wsUrl = HomeApp.utils.getWsUrl('/ws/monitor', 'id=' + monitorId);
                if (this.loadingComplete) {
                    // 如果已经加载过历史数据，添加 no_history 参数
                    wsUrl += '&no_history=1';
                }

                const handlers = {
                    onopen: function() {
                        console.log('Monitor WebSocket connected');
                    },
                    onmessage: function(event) {
                        try {
                            const data = JSON.parse(event.data);
                            if (data.type === 'monitor_detail') {
                                self.data = data.data;
                                self.render();
                            } else if (data.type === 'history_day' && !self.loadingComplete) {
                                // 只在历史数据未加载完成时才处理
                                self.historyDays.push(data);
                                self.renderHistory();
                            } else if (data.type === 'history_done' && !self.loadingComplete) {
                                // 只在历史数据未加载完成时才处理
                                self.loadingComplete = true;
                                self.renderHistoryComplete();
                            }
                        } catch (e) {
                            console.error('Failed to parse message:', e);
                        }
                    },
                    onclose: function() {
                        console.log('Monitor WebSocket disconnected, reconnecting...');
                        HomeApp.reconnect(wsUrl, handlers);
                    },
                    onerror: function(error) {
                        console.error('Monitor WebSocket error:', error);
                    }
                };
                HomeApp.connect(wsUrl, handlers);
            },

            render: function() {
                const container = document.getElementById('monitor-container');
                if (!container || !this.data) return;

                // 清理所有旧的提示元素
                HomeApp.clearTips();

                const data = this.data;

                let html = '';

                // 渲染状态卡片
                html += HomeApp.renderStatusCard(data, false);

                // 添加历史数据容器
                html += '<div id="history-container" class="history-container">';
                html += '<div id="history-loading" class="loading" style="text-align: center; padding: 40px 0; color: #7f8c8d;">加载中...</div>';
                html += '<div id="history-days"></div>';
                html += '</div>';

                container.innerHTML = html;
            },

            renderHistory: function() {
                const historyDaysEl = document.getElementById('history-days');
                if (!historyDaysEl) return;

                // 清空现有内容，重新渲染所有天
                historyDaysEl.innerHTML = '';

                // 去重并按日期排序
                const uniqueDays = {};
                this.historyDays.forEach(function(dayData) {
                    uniqueDays[dayData.date] = dayData;
                });

                // 转换为数组并按日期排序（从旧到新）
                const sortedDays = Object.values(uniqueDays).sort(function(a, b) {
                    return new Date(a.date) - new Date(b.date);
                });

                sortedDays.forEach(function(dayData) {
                    // 使用卡片式风格
                    let dayHtml = '<div class="status-card" style="margin-bottom: 20px;">';
                    dayHtml += '<div class="status-header">';
                    dayHtml += '<div class="service-name">最近7天监控数据 - ' + dayData.date + '</div>';
                    dayHtml += '<div class="status-indicator">';
                    dayHtml += '<span style="color: #27ae60;">正常: ' + dayData.up_count + '</span> / ';
                    dayHtml += '<span style="color: #e74c3c;">故障: ' + dayData.down_count + '</span> / ';
                    dayHtml += '可用率: ' + dayData.up_rate.toFixed(1) + '%';
                    dayHtml += '</div>';
                    dayHtml += '</div>';

                    // 渲染时间网格
                    dayHtml += '<div class="time-grid" style="max-height: none; overflow: visible;">';
                    dayData.logs.forEach(function(log) {
                        const status = log.is_valid ? 'up' : 'down';
                        const error = log.error_msg || '无法访问';
                        const timeStr = log.time || '';
                        let cellHtml = '<div class="time-cell ' + status + '" data-status="' + status + '" data-time="' + timeStr + '" data-error="' + HomeApp.utils.escapeHtml(error) + '"';
                        if (log.speed) {
                            cellHtml += ' data-speed="' + log.speed + '"';
                        }
                        if (log.size !== undefined && log.size !== null) {
                            let sizeStr;
                            if (log.size >= 1024) {
                                sizeStr = (log.size / 1024).toFixed(2) + 'kb';
                            } else {
                                sizeStr = log.size + 'b';
                            }
                            cellHtml += ' data-size="' + sizeStr + '"';
                        }
                        cellHtml += '></div>';
                        dayHtml += cellHtml;
                    });
                    dayHtml += '</div>';

                    // 渲染图例
                    dayHtml += HomeApp.renderLegend();

                    // 渲染页脚统计
                    dayHtml += '<div class="status-footer">';
                    dayHtml += '<div class="status-stats">日志: ' + dayData.total + ' 条</div>';
                    if (dayData.logs && dayData.logs.length > 0) {
                        const lastLog = dayData.logs[dayData.logs.length - 1];
                        if (lastLog.is_valid) {
                            if (lastLog.speed) {
                                dayHtml += '<div class="status-stats">耗时: ' + lastLog.speed + 'ms</div>';
                            }
                            if (lastLog.size !== undefined && lastLog.size !== null) {
                                let sizeStr;
                                if (lastLog.size >= 1024) {
                                    sizeStr = (lastLog.size / 1024).toFixed(2) + 'kb';
                                } else {
                                    sizeStr = lastLog.size + 'b';
                                }
                                dayHtml += '<div class="status-stats">大小: ' + sizeStr + '</div>';
                            }
                        }
                    }
                    dayHtml += '<div class="status-stats">可用率: ' + dayData.up_rate.toFixed(1) + '%</div>';
                    dayHtml += '</div>';

                    dayHtml += '</div>';
                    historyDaysEl.innerHTML += dayHtml;
                });
            },

            renderHistoryComplete: function() {
                const loadingEl = document.getElementById('history-loading');
                if (loadingEl) {
                    loadingEl.style.display = 'none';
                }

                if (this.historyDays.length === 0) {
                    const historyDaysEl = document.getElementById('history-days');
                    if (historyDaysEl) {
                        historyDaysEl.innerHTML = '<div style="text-align: center; padding: 40px 0; color: #999;">暂无7天数据</div>';
                    }
                }
            }
        },

        // 初始化事件处理
        initEventHandlers: function() {
            // 为时间格子添加鼠标悬停事件
            document.addEventListener('mouseenter', function(e) {
                if (!e.target || !e.target.classList) return;
                if (e.target.classList.contains('time-cell')) {
                    const cell = e.target;
                    const status = cell.dataset.status;
                    const time = cell.dataset.time;
                    const error = cell.dataset.error || '';

                    if (!cell._tip) {
                        // 创建提示元素
                        const tip = document.createElement('div');
                        tip.className = 'time-cell-tip';
                        tip.style.position = 'fixed';
                        tip.style.backgroundColor = '#333';
                        tip.style.color = '#fff';
                        tip.style.padding = '5px 10px';
                        tip.style.borderRadius = '4px';
                        tip.style.fontSize = '12px';
                        tip.style.zIndex = '10000';
                        tip.style.pointerEvents = 'none';
                        tip.style.maxWidth = '300px';
                        tip.style.wordBreak = 'break-word';

                        // 根据状态设置不同的内容和样式
                        if (status === 'up') {
                            tip.innerHTML = '状态: <span style="color: #4CAF50;">正常</span> | 时间: ' + time;
                            // 检查是否有耗时和大小信息
                            if (cell.dataset.speed) {
                                tip.innerHTML += ' | 耗时: ' + cell.dataset.speed + 'ms';
                            }
                            // 检查大小信息，包括0
                            if (cell.dataset.size !== undefined && cell.dataset.size !== null) {
                                tip.innerHTML += ' | 大小: ' + cell.dataset.size;
                            }
                        } else if (status === 'down') {
                            tip.innerHTML = '状态: <span style="color: #f44336;">故障</span> | 时间: ' + time + ' | 错误: <span style="color: #f44336;">' + error + '</span>';
                        } else {
                            tip.innerHTML = '状态: <span style="color: #ff9800;">未知</span> | 时间: ' + time;
                        }

                        document.body.appendChild(tip);

                        // 计算位置，确保tooltip在视口内
                        let top = e.clientY + 10;
                        let left = e.clientX + 10;
                        
                        const tipWidth = tip.offsetWidth;
                        const tipHeight = tip.offsetHeight;
                        const viewportWidth = window.innerWidth;
                        const viewportHeight = window.innerHeight;
                        
                        // 调整位置
                        if (left + tipWidth > viewportWidth) {
                            left = e.clientX - tipWidth - 10;
                        }
                        if (top + tipHeight > viewportHeight) {
                            top = e.clientY - tipHeight - 10;
                        }
                        
                        // 确保不超出视口
                        left = Math.max(10, left);
                        top = Math.max(10, top);

                        // 定位提示元素
                        tip.style.top = top + 'px';
                        tip.style.left = left + 'px';

                        // 存储提示元素引用
                        cell._tip = tip;
                    }
                }
            }, true);

            document.addEventListener('mouseleave', function(e) {
                if (!e.target || !e.target.classList) return;
                if (e.target.classList.contains('time-cell')) {
                    const cell = e.target;
                    if (cell._tip) {
                        document.body.removeChild(cell._tip);
                        delete cell._tip;
                    }
                }
            }, true);

            document.addEventListener('mousemove', function(e) {
                if (!e.target || !e.target.classList) return;
                if (e.target.classList.contains('time-cell') && e.target._tip) {
                    let top = e.clientY + 10;
                    let left = e.clientX + 10;
                    
                    const tip = e.target._tip;
                    const tipWidth = tip.offsetWidth;
                    const tipHeight = tip.offsetHeight;
                    const viewportWidth = window.innerWidth;
                    const viewportHeight = window.innerHeight;
                    
                    // 调整位置
                    if (left + tipWidth > viewportWidth) {
                        left = e.clientX - tipWidth - 10;
                    }
                    if (top + tipHeight > viewportHeight) {
                        top = e.clientY - tipHeight - 10;
                    }
                    
                    // 确保不超出视口
                    left = Math.max(10, left);
                    top = Math.max(10, top);
                    
                    tip.style.top = top + 'px';
                    tip.style.left = left + 'px';
                }
            }, true);

            // 为状态指示器添加鼠标悬停事件
            document.addEventListener('mouseenter', function(e) {
                if (!e.target || !e.target.closest) return;
                const indicator = e.target.closest('.status-indicator');
                if (indicator) {
                    const card = indicator.closest('.status-card');
                    const errorTooltip = card ? card.querySelector('.error-tooltip') : null;
                    const error = errorTooltip ? errorTooltip.dataset.error : '';

                    if (error && !indicator._tip) {
                        // 创建提示元素
                        const tip = document.createElement('div');
                        tip.className = 'status-indicator-tip';
                        tip.style.position = 'fixed';
                        tip.style.backgroundColor = '#333';
                        tip.style.color = '#fff';
                        tip.style.padding = '5px 10px';
                        tip.style.borderRadius = '4px';
                        tip.style.fontSize = '12px';
                        tip.style.zIndex = '10000';
                        tip.style.pointerEvents = 'none';
                        tip.style.whiteSpace = 'pre-wrap';
                        tip.style.maxWidth = '300px';

                        tip.innerHTML = '<span style="color: #f44336;">错误信息:</span><br>' + error;

                        document.body.appendChild(tip);

                        // 定位提示元素
                        tip.style.top = (e.clientY + 10) + 'px';
                        tip.style.left = (e.clientX + 10) + 'px';

                        // 存储提示元素引用
                        indicator._tip = tip;
                    }
                }
            }, true);

            document.addEventListener('mouseleave', function(e) {
                if (!e.target || !e.target.closest) return;
                const indicator = e.target.closest('.status-indicator');
                if (indicator) {
                    if (indicator._tip) {
                        document.body.removeChild(indicator._tip);
                        delete indicator._tip;
                    }
                }
            }, true);

            document.addEventListener('mousemove', function(e) {
                if (!e.target || !e.target.closest) return;
                const indicator = e.target.closest('.status-indicator');
                if (indicator && indicator._tip) {
                    indicator._tip.style.top = (e.clientY + 10) + 'px';
                    indicator._tip.style.left = (e.clientX + 10) + 'px';
                }
            }, true);
        },

        // 初始化应用
        init: function() {
            // 初始化事件处理
            this.initEventHandlers();

            // 根据页面类型初始化不同的渲染器
            const statusContainer = document.getElementById('status-container');
            if (statusContainer) {
                this.renderIndex.init();
                return;
            }

            const groupContainer = document.getElementById('group-container');
            if (groupContainer) {
                this.renderGroups.init();
                return;
            }

            const monitorContainer = document.getElementById('monitor-container');
            if (monitorContainer) {
                this.renderMonitor.init();
            }
        }
    };

    document.addEventListener('DOMContentLoaded', function() {
        HomeApp.init();
    });

    window.HomeApp = HomeApp;

})();