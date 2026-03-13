(() => {
    'use strict';

    const SELECTORS = {
        contentArea: '.content-area',
        homeLink: '[data-nav="home"]',
        inventoryItem: '.inventory-item'
    };

    const API = {
        featuredSkin: '/api/home/featured-skin',
        topMovers: '/api/home/top-movers',
        skinPrice: '/api/skin-price/'
    };

    let currentChart = null;

    function getContentArea() {
        return document.querySelector(SELECTORS.contentArea);
    }

    function clearChart() {
        if (!currentChart) {
            return;
        }

        currentChart.destroy();
        currentChart = null;
    }

    function showMessage(contentArea, message) {
        clearChart();
        contentArea.innerHTML = `<p class="page-message">${message}</p>`;
    }

    function normalizeHistory(rawHistory) {
        if (!Array.isArray(rawHistory)) {
            return [];
        }

        return rawHistory
            .map((item) => ({
                date: item?.date,
                price: Number(item?.price)
            }))
            .filter((item) => item.date && Number.isFinite(item.price));
    }

    function normalizeMovers(rawMovers) {
        if (!Array.isArray(rawMovers)) {
            return [];
        }

        return rawMovers
            .map((item) => ({
                skinName: item?.skinName ?? item?.name ?? 'Unknown skin',
                change24h: Number(item?.change24h ?? item?.change ?? 0),
                price: Number(item?.price ?? 0)
            }))
            .filter((item) => item.skinName && Number.isFinite(item.change24h))
            .sort((a, b) => Math.abs(b.change24h) - Math.abs(a.change24h))
            .slice(0, 5);
    }

    function readFeaturedModel(payload) {
        const history = normalizeHistory(payload?.history ?? payload?.priceData ?? payload?.prices);

        return {
            skinId: payload?.skinId ?? payload?.id ?? '',
            skinName: payload?.skinName ?? payload?.name ?? 'Unknown skin',
            currentPrice: Number(payload?.currentPrice ?? payload?.price ?? history.at(-1)?.price ?? 0),
            change7d: Number(payload?.change7d ?? payload?.weekChange ?? 0),
            change30d: Number(payload?.change30d ?? payload?.monthChange ?? 0),
            updatedAt: payload?.updatedAt ?? payload?.lastUpdated ?? new Date().toISOString(),
            movers: normalizeMovers(payload?.movers ?? payload?.topMovers),
            history
        };
    }

    async function fetchJson(url) {
        const response = await fetch(url);

        if (!response.ok) {
            throw new Error(`Request failed with status ${response.status}`);
        }

        return response.json();
    }

    async function fetchFirstAvailableJson(urls) {
        const failures = [];

        for (const url of urls) {
            try {
                const payload = await fetchJson(url);
                return { payload, sourceUrl: url };
            } catch (error) {
                failures.push(`${url} -> ${error.message}`);
            }
        }

        throw new Error(failures.join(' | '));
    }

    function getFeaturedSkinUrls() {
        const configured = window.APP_CONFIG?.featuredSkinUrl;
        const candidates = [
            configured,
            API.featuredSkin,
            '/api/featured-skin',
            '/home/featured-skin'
        ].filter(Boolean);

        return [...new Set(candidates)];
    }

    function getTopMoversUrls() {
        const configured = window.APP_CONFIG?.topMoversUrl;
        const candidates = [
            configured,
            API.topMovers,
            '/api/top-movers',
            '/home/top-movers'
        ].filter(Boolean);

        return [...new Set(candidates)];
    }

    async function fetchFeaturedSkin() {
        const urls = getFeaturedSkinUrls();
        const { payload } = await fetchFirstAvailableJson(urls);
        return readFeaturedModel(payload);
    }

    async function fetchTopMovers() {
        const urls = getTopMoversUrls();
        const { payload } = await fetchFirstAvailableJson(urls);
        return normalizeMovers(payload?.movers ?? payload?.topMovers ?? payload);
    }

    function buildMockFeaturedSkin() {
        const today = new Date();
        const history = [];
        let price = 148.5;

        for (let i = 29; i >= 0; i -= 1) {
            const d = new Date(today);
            d.setDate(today.getDate() - i);
            price += (Math.random() - 0.45) * 1.6;
            history.push({
                date: d.toISOString().slice(0, 10),
                price: Math.max(120, Number(price.toFixed(2)))
            });
        }

        return {
            skinId: 'demo-featured-skin',
            skinName: 'Demo Featured Skin',
            currentPrice: history.at(-1).price,
            change7d: 2.14,
            change30d: 6.73,
            updatedAt: new Date().toISOString(),
            movers: buildMockTopMovers(),
            history
        };
    }

    function buildMockTopMovers() {
        return [
            { skinName: 'AK-47 | Redline', change24h: 3.42, price: 124.6 },
            { skinName: 'AWP | Asiimov', change24h: -2.15, price: 198.1 },
            { skinName: 'M4A4 | Neo-Noir', change24h: 4.03, price: 72.4 },
            { skinName: 'USP-S | Kill Confirmed', change24h: -1.56, price: 88.9 },
            { skinName: 'Desert Eagle | Printstream', change24h: 2.38, price: 93.7 }
        ];
    }

    async function fetchSkinPriceData(skinId) {
        const encodedId = encodeURIComponent(skinId);
        const payload = await fetchJson(`${API.skinPrice}${encodedId}`);

        return {
            skinId,
            skinName: skinId,
            currentPrice: 0,
            change7d: 0,
            change30d: 0,
            history: normalizeHistory(payload)
        };
    }

    function formatPrice(value) {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD',
            maximumFractionDigits: 2
        }).format(value);
    }

    function formatChange(value) {
        const sign = value > 0 ? '+' : '';
        return `${sign}${value.toFixed(2)}%`;
    }

    function formatDateTime(value) {
        const date = new Date(value);

        if (Number.isNaN(date.getTime())) {
            return 'Unknown time';
        }

        return new Intl.DateTimeFormat('en-US', {
            month: 'short',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        }).format(date);
    }

    function changeClass(value) {
        if (value > 0) {
            return 'positive';
        }

        if (value < 0) {
            return 'negative';
        }

        return '';
    }

    function renderMovers(movers) {
        if (!movers.length) {
            return '<li class="movers-empty">No movers data yet.</li>';
        }

        return movers
            .map((item) => `
                <li class="movers-item">
                    <span class="movers-name">${item.skinName}</span>
                    <span class="movers-change ${changeClass(item.change24h)}">${formatChange(item.change24h)}</span>
                </li>
            `)
            .join('');
    }

    function renderHomeLayout(contentArea, skin) {
        contentArea.innerHTML = `
            <section class="home-panel">
                <header class="home-header">
                    <div>
                        <h3>${skin.skinName}</h3>
                        <p class="home-subtitle">Most expensive skin from your market data</p>
                    </div>
                    <p class="updated-at">Updated: ${formatDateTime(skin.updatedAt)}</p>
                </header>

                <section class="metrics">
                    <article class="metric">
                        <p class="metric-label">Current price</p>
                        <p class="metric-value">${formatPrice(skin.currentPrice)}</p>
                    </article>
                    <article class="metric">
                        <p class="metric-label">7d change</p>
                        <p class="metric-value ${changeClass(skin.change7d)}">${formatChange(skin.change7d)}</p>
                    </article>
                    <article class="metric">
                        <p class="metric-label">30d change</p>
                        <p class="metric-value ${changeClass(skin.change30d)}">${formatChange(skin.change30d)}</p>
                    </article>
                </section>

                <section class="home-grid">
                    <div class="chart-shell">
                        <canvas id="homeFeaturedChart"></canvas>
                    </div>
                    <aside class="movers-panel">
                        <h4>Top movers (24h)</h4>
                        <ul class="movers-list">${renderMovers(skin.movers ?? [])}</ul>
                    </aside>
                </section>
            </section>
        `;

        return contentArea.querySelector('#homeFeaturedChart');
    }

    function renderLineChart(canvas, label, points) {
        clearChart();

        currentChart = new Chart(canvas, {
            type: 'line',
            data: {
                labels: points.map((point) => point.date),
                datasets: [
                    {
                        label,
                        data: points.map((point) => point.price),
                        borderColor: 'rgb(189, 189, 189)',
                        backgroundColor: 'rgba(189, 189, 189, 0.12)',
                        borderWidth: 2,
                        tension: 0.35,
                        fill: true
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        labels: {
                            color: 'rgb(189, 189, 189)'
                        }
                    }
                },
                scales: {
                    y: {
                        ticks: {
                            color: 'rgb(189, 189, 189)'
                        },
                        grid: {
                            color: 'rgba(60, 60, 60, 0.5)'
                        }
                    },
                    x: {
                        ticks: {
                            color: 'rgb(189, 189, 189)'
                        },
                        grid: {
                            color: 'rgba(60, 60, 60, 0.5)'
                        }
                    }
                }
            }
        });
    }

    async function loadHome(contentArea) {
        showMessage(contentArea, 'Loading home data...');

        try {
            const skin = await fetchFeaturedSkin();

            if (!Array.isArray(skin.movers) || skin.movers.length === 0) {
                try {
                    skin.movers = await fetchTopMovers();
                } catch (error) {
                    skin.movers = buildMockTopMovers();
                }
            }

            if (skin.history.length === 0) {
                showMessage(contentArea, 'No historical data for featured skin.');
                return;
            }

            if (!Number.isFinite(skin.currentPrice) || skin.currentPrice <= 0) {
                skin.currentPrice = skin.history.at(-1).price;
            }

            const chartCanvas = renderHomeLayout(contentArea, skin);
            renderLineChart(chartCanvas, `Price ${skin.skinName}`, skin.history);
        } catch (error) {
            console.error('Failed to load home featured skin:', error);

            const demo = buildMockFeaturedSkin();
            const chartCanvas = renderHomeLayout(contentArea, demo);
            renderLineChart(chartCanvas, `Price ${demo.skinName}`, demo.history);
        }
    }

    async function loadSkinFromInventory(item, contentArea) {
        const skinId = item.dataset.skinId;
        const skinName = item.dataset.skinName || skinId || 'Unknown skin';

        if (!skinId) {
            showMessage(contentArea, 'Selected inventory item has no skinId.');
            return;
        }

        showMessage(contentArea, `Loading ${skinName}...`);

        try {
            const model = await fetchSkinPriceData(skinId);
            model.skinName = skinName;

            if (model.history.length === 0) {
                showMessage(contentArea, `No price history for ${skinName}.`);
                return;
            }

            model.currentPrice = model.history.at(-1).price;
            const chartCanvas = renderHomeLayout(contentArea, model);
            renderLineChart(chartCanvas, `Price ${skinName}`, model.history);
        } catch (error) {
            console.error('Failed to load inventory skin chart:', error);
            showMessage(contentArea, `Failed to load ${skinName}.`);
        }
    }

    function initHomeNav(contentArea) {
        const homeLink = document.querySelector(SELECTORS.homeLink);

        if (!homeLink) {
            return;
        }

        homeLink.addEventListener('click', (event) => {
            event.preventDefault();
            void loadHome(contentArea);
        });
    }

    function initInventoryClicks(contentArea) {
        document.addEventListener('click', (event) => {
            const item = event.target.closest(SELECTORS.inventoryItem);

            if (!item) {
                return;
            }

            void loadSkinFromInventory(item, contentArea);
        });
    }

    function init() {
        const contentArea = getContentArea();

        if (!contentArea) {
            console.error('Missing .content-area in HTML.');
            return;
        }

        initHomeNav(contentArea);
        initInventoryClicks(contentArea);
        void loadHome(contentArea);
    }

    document.addEventListener('DOMContentLoaded', init);
})();
