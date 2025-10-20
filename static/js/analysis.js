let priceChart = null;
let currentData = null;
let currentChartType = 'line';
let originalData = null;
let visibleStart = 0;
let visibleEnd = 0;

const pairSelect = document.getElementById('pairSelect');
const timeframeSelect = document.getElementById('timeframeSelect');
const currentPairEl = document.getElementById('currentPair');
const currentTimeframeEl = document.getElementById('currentTimeframe');
const indicatorsContainer = document.getElementById('indicatorsContainer');
const errorContainer = document.getElementById('errorContainer');
const lastUpdateEl = document.getElementById('lastUpdate');

document.addEventListener('DOMContentLoaded', () => {
    setupEventListeners();
    loadData();
});

function setupEventListeners() {
    pairSelect.addEventListener('change', loadData);
    timeframeSelect.addEventListener('change', loadData);

    document.querySelectorAll('.chart-btn[data-type]').forEach(btn => {
        btn.addEventListener('click', (e) => {
            document.querySelectorAll('.chart-btn[data-type]').forEach(b => b.classList.remove('active'));
            e.target.classList.add('active');
            currentChartType = e.target.dataset.type;
            if (currentData) updatePriceChart(currentData);
        });
    });

    document.getElementById('btnZoomReset').addEventListener('click', () => {
        if (originalData) {
            visibleStart = 0;
            visibleEnd = originalData.labels.length - 1;
            updateVisibleRange(priceChart, originalData, visibleStart, visibleEnd);
        }
    });
}

async function loadData() {
    const pair = pairSelect.value;
    const timeframe = timeframeSelect.value;

    currentPairEl.textContent = pair.replace('USDT', '/USDT');
    currentTimeframeEl.textContent = `(${timeframe})`;

    showLoading();
    hideError();

    try {
        const response = await fetch(`/api/pair?pair=${pair}&timeframe=${timeframe}`);
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        const data = await response.json();

        currentData = data;
        updateDashboard(data);
        updateLastUpdate();
    } catch (err) {
        showError('Не удалось загрузить данные: ' + err.message);
        console.error(err);
    }
}

function updateDashboard(data) {
    updatePriceChart(data);
    updateIndicators(data);
}

function updatePriceChart(data) {
    const canvas = document.getElementById('priceChart');
    const ctx = canvas.getContext('2d');

    if (priceChart) {
        priceChart.destroy();
    }

    const candles = data.candles || [];
    if (candles.length === 0) {
        showError('Нет данных для отображения графика');
        return;
    }

    const labels = candles.map(candle => {
        const date = new Date(candle.openTime);
        return date.toLocaleString('ru-RU', {
            day: 'numeric',
            month: 'short',
            hour: '2-digit',
            minute: '2-digit'
        });
    });

    let datasets = [];

    if (currentChartType === 'ohlc') {
        datasets = [
            {
                label: 'High',
                data: candles.map(c => c.high),
                borderColor: '#f0b90b',
                backgroundColor: 'rgba(240, 185, 11, 0.1)',
                borderWidth: 1,
                pointRadius: 0
            },
            {
                label: 'Low',
                data: candles.map(c => c.low),
                borderColor: '#f6465d',
                backgroundColor: 'rgba(246, 70, 93, 0.1)',
                borderWidth: 1,
                pointRadius: 0
            },
            {
                label: 'Close',
                data: candles.map(c => c.close),
                borderColor: '#3b82f6',
                backgroundColor: 'rgba(59, 130, 246, 0.1)',
                borderWidth: 2,
                pointRadius: 0
            }
        ];
    } else {
        datasets = [{
            label: 'Цена закрытия',
            data: candles.map(c => c.close),
            borderColor: '#f0b90b',
            backgroundColor: 'rgba(240, 185, 11, 0.1)',
            borderWidth: 2,
            fill: true,
            tension: 0.1,
            pointRadius: 0
        }];
    }

    // Сохраняем оригинальные данные для управления видимой областью
    originalData = {
        labels: labels,
        datasets: datasets.map(dataset => ({
            ...dataset,
            originalData: [...dataset.data]
        }))
    };

    // Начальная видимая область - показываем все данные
    visibleStart = 0;
    visibleEnd = labels.length - 1;

    priceChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: originalData.labels.slice(visibleStart, visibleEnd + 1),
            datasets: originalData.datasets.map(dataset => ({
                ...dataset,
                data: dataset.originalData.slice(visibleStart, visibleEnd + 1)
            }))
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: currentChartType === 'ohlc',
                    labels: { color: '#b7b7b7' }
                },
                tooltip: {
                    mode: 'index',
                    intersect: false,
                    callbacks: {
                        label: function (context) {
                            const label = context.dataset.label || '';
                            const value = context.parsed.y;
                            return `${label}: $${value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
                        }
                    }
                }
            },
            scales: {
                x: {
                    ticks: {
                        color: '#b7b7b7',
                        maxTicksLimit: 10,
                        font: { size: 11 }
                    },
                    grid: { color: 'rgba(183, 183, 183, 0.1)' }
                },
                y: {
                    ticks: {
                        color: '#b7b7b7',
                        callback: v => '$' + v.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
                        font: { size: 11 }
                    },
                    grid: { color: 'rgba(183, 183, 183, 0.1)' }
                }
            },
            interaction: { intersect: false, mode: 'index' },
            elements: {
                point: { radius: 0, hoverRadius: 4 },
                line: { tension: currentChartType === 'line' ? 0.1 : 0 }
            }
        }
    });

    // Добавляем функцию перетаскивания
    addDragToPan(canvas, priceChart, originalData);
}

// Функция для добавления перетаскивания
function addDragToPan(canvas, chart, data) {
    let isDragging = false;
    let startX = 0;
    let startVisibleStart = visibleStart;
    let startVisibleEnd = visibleEnd;

    // Для мультитач зума
    let initialPinchDistance = 0;
    let initialVisibleRange = 0;
    let isPinching = false;
    let lastZoomTime = 0;
    const ZOOM_THROTTLE_MS = 16; // ~60fps для плавности

    canvas.style.cursor = 'grab';

    // === ОБРАБОТЧИКИ МЫШИ ===
    canvas.addEventListener('mousedown', handleMouseDown);
    canvas.addEventListener('mousemove', handleMouseMove);
    canvas.addEventListener('mouseup', handleMouseUp);
    canvas.addEventListener('mouseleave', handleMouseUp);
    canvas.addEventListener('wheel', handleWheel);

    // === ОБРАБОТЧИКИ ТАЧ-СОБЫТИЙ ===
    canvas.addEventListener('touchstart', handleTouchStart, { passive: false });
    canvas.addEventListener('touchmove', handleTouchMove, { passive: false });
    canvas.addEventListener('touchend', handleTouchEnd);
    canvas.addEventListener('touchcancel', handleTouchEnd);

    function handleMouseDown(e) {
        startDragging(e.clientX);
        e.preventDefault();
    }

    function handleMouseMove(e) {
        if (!isDragging) return;
        const deltaX = e.clientX - startX;
        updatePanPosition(deltaX);
    }

    function handleMouseUp() {
        stopDragging();
    }

    function handleWheel(e) {
        e.preventDefault();
        // Более плавный зум для колесика
        const zoomFactor = e.deltaY > 0 ? 0.9 : 1.1;
        handleZoom(zoomFactor, e.clientX);
    }

    // === ТАЧ-ФУНКЦИИ ===
    function handleTouchStart(e) {
        if (e.touches.length === 1) {
            startDragging(e.touches[0].clientX);
            e.preventDefault();
        } else if (e.touches.length === 2) {
            startPinching(e.touches[0], e.touches[1]);
            e.preventDefault();
        }
    }

    function handleTouchMove(e) {
        if (e.touches.length === 1 && isDragging) {
            const deltaX = e.touches[0].clientX - startX;
            updatePanPosition(deltaX);
            e.preventDefault();
        } else if (e.touches.length === 2 && isPinching) {
            handlePinchZoom(e.touches[0], e.touches[1]);
            e.preventDefault();
        }
    }

    function handleTouchEnd(e) {
        if (e.touches.length === 0) {
            stopDragging();
            resetPinch();
        } else if (e.touches.length === 1) {
            stopDragging();
            startDragging(e.touches[0].clientX);
        }
    }

    // === ОСНОВНЫЕ ФУНКЦИИ ===
    function startDragging(clientX) {
        isDragging = true;
        startX = clientX;
        startVisibleStart = visibleStart;
        startVisibleEnd = visibleEnd;
        canvas.style.cursor = 'grabbing';
    }

    function stopDragging() {
        isDragging = false;
        canvas.style.cursor = 'grab';
    }

    function updatePanPosition(deltaX) {
        const totalVisiblePoints = startVisibleEnd - startVisibleStart;
        const movePoints = Math.round((deltaX / canvas.offsetWidth) * totalVisiblePoints);

        visibleStart = Math.max(0, startVisibleStart - movePoints);
        visibleEnd = Math.min(data.labels.length - 1, startVisibleEnd - movePoints);

        if (visibleEnd - visibleStart !== totalVisiblePoints) {
            if (visibleStart === 0) {
                visibleEnd = Math.min(data.labels.length - 1, totalVisiblePoints);
            } else if (visibleEnd === data.labels.length - 1) {
                visibleStart = Math.max(0, data.labels.length - 1 - totalVisiblePoints);
            }
        }

        updateVisibleRange(chart, data, visibleStart, visibleEnd);
    }

    function handleZoom(zoomFactor, centerX) {
        const now = Date.now();
        if (now - lastZoomTime < ZOOM_THROTTLE_MS) return;
        lastZoomTime = now;

        const rect = canvas.getBoundingClientRect();
        const relativeX = (centerX - rect.left) / rect.width;
        const centerIndex = Math.round(visibleStart + (visibleEnd - visibleStart) * relativeX);

        const currentRange = visibleEnd - visibleStart;
        const newRange = Math.round(currentRange / zoomFactor);

        const minRange = 5;
        const totalDataPoints = data.labels.length;

        // Разрешаем увеличивать до полного диапазона данных
        if (newRange >= minRange && newRange <= totalDataPoints) {
            let newStart = Math.max(0, centerIndex - Math.floor(newRange * relativeX));
            let newEnd = Math.min(totalDataPoints - 1, newStart + newRange);

            // Если пытаемся показать больше чем есть данных - показываем всё
            if (newRange >= totalDataPoints - 1) {
                newStart = 0;
                newEnd = totalDataPoints - 1;
            } else {
                // Корректируем границы
                if (newEnd > totalDataPoints - 1) {
                    newEnd = totalDataPoints - 1;
                    newStart = Math.max(0, newEnd - newRange);
                } else if (newStart < 0) {
                    newStart = 0;
                    newEnd = Math.min(totalDataPoints - 1, newRange);
                }
            }

            requestAnimationFrame(() => {
                visibleStart = newStart;
                visibleEnd = newEnd;
                updateVisibleRange(chart, data, visibleStart, visibleEnd);
            });
        }
    }

    // === ФУНКЦИИ ДЛЯ МУЛЬТИТАЧ ЗУМА ===
    function startPinching(touch1, touch2) {
        isPinching = true;
        initialPinchDistance = getDistance(touch1, touch2);
        initialVisibleRange = visibleEnd - visibleStart;
        lastZoomTime = Date.now();
    }

    function handlePinchZoom(touch1, touch2) {
        const now = Date.now();
        if (now - lastZoomTime < ZOOM_THROTTLE_MS) return;

        const currentDistance = getDistance(touch1, touch2);

        // Более плавный зум с небольшим коэффициентом
        const zoomFactor = currentDistance / initialPinchDistance;

        const centerX = (touch1.clientX + touch2.clientX) / 2;

        // Более плавные ограничения
        const constrainedZoomFactor = Math.max(0.8, Math.min(1.2, zoomFactor));

        handleZoom(constrainedZoomFactor, centerX);

        // Обновляем для плавности
        initialPinchDistance = currentDistance;
        lastZoomTime = now;
    }

    function resetPinch() {
        isPinching = false;
        initialPinchDistance = 0;
        initialVisibleRange = 0;
    }

    function getDistance(touch1, touch2) {
        const dx = touch1.clientX - touch2.clientX;
        const dy = touch1.clientY - touch2.clientY;
        return Math.sqrt(dx * dx + dy * dy);
    }
}

// Функция для обновления видимой области графика
function updateVisibleRange(chart, data, start, end) {
    chart.data.labels = data.labels.slice(start, end + 1);

    chart.data.datasets.forEach((dataset, index) => {
        dataset.data = data.datasets[index].originalData.slice(start, end + 1);
    });

    chart.update('none');
}

function updateIndicators(data) {
    const ind = data.indicators || {};
    const container = indicatorsContainer;

    // Получаем текущую цену из последней свечи
    const currentPrice = data.candles && data.candles.length > 0
        ? data.candles[data.candles.length - 1].close
        : 0;

    const rsiSignal = () => {
        if (ind.rsi == null || ind.rsi === 0) return { text: 'Нет данных', cls: 'signal-neutral', icon: '⚪' };
        if (ind.rsi > 70) return { text: 'Перекупленность', cls: 'signal-bearish', icon: '🔴' };
        if (ind.rsi < 30) return { text: 'Перепроданность', cls: 'signal-bullish', icon: '🟢' };
        return { text: 'Нейтрально', cls: 'signal-neutral', icon: '⚪' };
    };

    const macdSignal = () => {
        if (ind.macd == null || ind.signal == null) return { text: 'Нет данных', cls: 'signal-neutral', icon: '⚪' };
        if (ind.macd > ind.signal) return { text: 'Бычий сигнал', cls: 'signal-bullish', icon: '🟢' };
        return { text: 'Медвежий сигнал', cls: 'signal-bearish', icon: '🔴' };
    };

    const smaSignal = () => {
        if (ind.sma20 == null || ind.sma50 == null)
            return { text: 'Недостаточно данных', cls: 'signal-neutral', icon: '⚪' };
        return ind.sma20 > ind.sma50
            ? { text: 'Бычий тренд', cls: 'signal-bullish', icon: '🟢' }
            : { text: 'Медвежий тренд', cls: 'signal-bearish', icon: '🔴' };
    };

    const emaSignal = () => {
        if (ind.ema12 == null || ind.ema26 == null)
            return { text: 'Недостаточно данных', cls: 'signal-neutral', icon: '⚪' };
        return ind.ema12 > ind.ema26
            ? { text: 'Бычий тренд', cls: 'signal-bullish', icon: '🟢' }
            : { text: 'Медвежий тренд', cls: 'signal-bearish', icon: '🔴' };
    };

    const rsiCls = (ind.rsi > 70) ? 'price-negative' : (ind.rsi < 30) ? 'price-positive' : '';
    const macdCls = (ind.macd > ind.signal) ? 'price-positive' : 'price-negative';

    container.innerHTML = `
                <div class="indicators-grid">
                    <div class="indicator-item">
                        <div class="indicator-content">
                            <div class="indicator-header">
                                <div class="indicator-name">💰 Текущая цена</div>
                            </div>
                            <div class="indicator-value">
                                $${currentPrice.toFixed(2)}
                            </div>
                            <div class="indicator-details">
                                Последнее значение закрытия
                            </div>
                            <div class="signal-container">
                                <div class="indicator-signal signal-neutral">
                                    ⚪ Актуально
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="indicator-item">
                        <div class="indicator-content">
                            <div class="indicator-header">
                                <div class="indicator-name">📊 Скользящие средние</div>
                            </div>
                            <div class="indicator-value">
                                <div class="sma-values">SMA 20: $${ind.sma20?.toFixed(2) || 'N/A'}</div>
                                <div class="sma-values">SMA 50: $${ind.sma50?.toFixed(2) || 'N/A'}</div>
                            </div>
                            <div class="signal-container">
                                <div class="indicator-signal ${smaSignal().cls}">
                                    ${smaSignal().icon} ${smaSignal().text}
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="indicator-item">
                        <div class="indicator-content">
                            <div class="indicator-header">
                                <div class="indicator-name">📈 EMA</div>
                            </div>
                            <div class="indicator-value">
                                <div class="sma-values">EMA 12: $${ind.ema12?.toFixed(2) || 'N/A'}</div>
                                <div class="sma-values">EMA 26: $${ind.ema26?.toFixed(2) || 'N/A'}</div>
                            </div>
                            <div class="signal-container">
                                <div class="indicator-signal ${emaSignal().cls}">
                                    ${emaSignal().icon} ${emaSignal().text}
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="indicator-item">
                        <div class="indicator-content">
                            <div class="indicator-header">
                                <div class="indicator-name">⚡ RSI (14)</div>
                            </div>
                            <div class="indicator-value ${rsiCls}">
                                ${(ind.rsi !== undefined && ind.rsi !== 0) ? ind.rsi.toFixed(2) : 'N/A'}
                            </div>
                            <div class="indicator-details">
                                Моментум индикатор
                            </div>
                            <div class="signal-container">
                                <div class="indicator-signal ${rsiSignal().cls}">
                                    ${rsiSignal().icon} ${rsiSignal().text}
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="indicator-item">
                        <div class="indicator-content">
                            <div class="indicator-header">
                                <div class="indicator-name">📊 MACD</div>
                            </div>
                            <div class="indicator-value">
                                <div class="sma-values">MACD: ${ind.macd?.toFixed(4) || 'N/A'}</div>
                                <div class="sma-values">Signal: ${ind.signal?.toFixed(4) || 'N/A'}</div>
                                <div class="sma-values">Histogram: ${ind.histogram?.toFixed(4) || 'N/A'}</div>
                            </div>
                            <div class="signal-container">
                                <div class="indicator-signal ${macdSignal().cls}">
                                    ${macdSignal().icon} ${macdSignal().text}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            `;
}

function showLoading() {
    indicatorsContainer.innerHTML = '<div class="loading"><i class="fas fa-spinner fa-spin"></i><br>Загрузка данных...</div>';
}

function showError(msg) {
    errorContainer.innerHTML = `<div class="error"><i class="fas fa-exclamation-triangle"></i> ${msg}</div>`;
    errorContainer.style.display = 'block';
}

function hideError() {
    errorContainer.style.display = 'none';
}

function updateLastUpdate() {
    lastUpdateEl.textContent = new Date().toLocaleString('ru-RU');
}