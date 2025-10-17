let chart = null;
let candleSeries = null;
let volumeSeries = null;
let sma20Series = null;
let sma50Series = null;

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', function () {
    initializeChart();
    setupEventListeners();

    // Автоматическая загрузка данных для BTC/USDT
    loadAnalysisData();
});

function initializeChart() {
    const chartContainer = document.getElementById('candleChart');

    // Создаем график
    chart = LightweightCharts.createChart(chartContainer, {
        width: chartContainer.clientWidth,
        height: 500,
        layout: {
            background: { color: 'transparent' },
            textColor: '#d1d4dc',
        },
        grid: {
            vertLines: { color: 'rgba(42, 46, 57, 0.5)' },
            horzLines: { color: 'rgba(42, 46, 57, 0.5)' },
        },
        crosshair: {
            mode: LightweightCharts.CrosshairMode.Normal,
        },
        rightPriceScale: {
            borderColor: 'rgba(197, 203, 206, 0.8)',
        },
        timeScale: {
            borderColor: 'rgba(197, 203, 206, 0.8)',
            timeVisible: true,
            secondsVisible: false,
        },
    });

    // Серия свечей
    candleSeries = chart.addCandlestickSeries({
        upColor: '#26a69a',
        downColor: '#ef5350',
        borderDownColor: '#ef5350',
        borderUpColor: '#26a69a',
        wickDownColor: '#ef5350',
        wickUpColor: '#26a69a',
    });

    // Серия объема
    volumeSeries = chart.addHistogramSeries({
        color: '#26a69a',
        priceFormat: {
            type: 'volume',
        },
        priceScaleId: '', // отдельная шкала
        scaleMargins: {
            top: 0.8,
            bottom: 0,
        },
    });

    // SMA 20
    sma20Series = chart.addLineSeries({
        color: 'rgba(4, 111, 232, 1)',
        lineWidth: 2,
        title: 'SMA 20',
    });

    // SMA 50
    sma50Series = chart.addLineSeries({
        color: 'rgba(245, 166, 35, 1)',
        lineWidth: 2,
        title: 'SMA 50',
    });

    // Адаптация к размеру окна
    window.addEventListener('resize', function () {
        chart.applyOptions({
            width: chartContainer.clientWidth,
        });
    });
}

function setupEventListeners() {
    document.getElementById('loadDataBtn').addEventListener('click', loadAnalysisData);

    // Загрузка при изменении пары или таймфрейма
    document.getElementById('pairSelect').addEventListener('change', loadAnalysisData);
    document.getElementById('timeframeSelect').addEventListener('change', loadAnalysisData);
}

async function loadAnalysisData() {
    const pair = document.getElementById('pairSelect').value;
    const timeframe = document.getElementById('timeframeSelect').value;
    const useCache = document.getElementById('useCache').checked;

    const loadBtn = document.getElementById('loadDataBtn');
    const originalText = loadBtn.textContent;

    try {
        // Показываем загрузку
        loadBtn.innerHTML = '<div class="loading"></div> Загрузка...';
        loadBtn.disabled = true;

        const response = await fetch('/api/analysis-data', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                pair: pair,
                timeframe: timeframe,
                useCache: useCache,
            }),
        });

        const data = await response.json();

        if (data.success) {
            updateChart(data.data);
            updateIndicators(data.data.indicators);
            updateAnalysisText(data.data);
        } else {
            throw new Error(data.error);
        }

    } catch (error) {
        console.error('Error loading analysis data:', error);
        showError('Ошибка при загрузке данных: ' + error.message);
    } finally {
        // Восстанавливаем кнопку
        loadBtn.innerHTML = originalText;
        loadBtn.disabled = false;
    }
}

function updateChart(analysisData) {
    // Подготавливаем данные для графика
    const candleData = analysisData.candles.map(candle => ({
        time: Math.floor(candle.openTime / 1000), // конвертируем в секунды
        open: candle.open,
        high: candle.high,
        low: candle.low,
        close: candle.close,
    }));

    const volumeData = analysisData.candles.map(candle => ({
        time: Math.floor(candle.openTime / 1000),
        value: candle.volume,
        color: candle.close >= candle.open ? 'rgba(38, 166, 154, 0.8)' : 'rgba(239, 83, 80, 0.8)',
    }));

    // Берем последние 100 свечей для SMA (чтобы не перегружать график)
    const recentCandles = analysisData.candles.slice(-100);
    const sma20Data = recentCandles.map((candle, index, array) => {
        if (index < 19) return null; // Нужно минимум 20 свечей для SMA20

        const start = Math.max(0, index - 19);
        const slice = array.slice(start, index + 1);
        const sum = slice.reduce((acc, c) => acc + c.close, 0);
        const sma = sum / slice.length;

        return {
            time: Math.floor(candle.openTime / 1000),
            value: sma,
        };
    }).filter(item => item !== null);

    const sma50Data = recentCandles.map((candle, index, array) => {
        if (index < 49) return null; // Нужно минимум 50 свечей для SMA50

        const start = Math.max(0, index - 49);
        const slice = array.slice(start, index + 1);
        const sum = slice.reduce((acc, c) => acc + c.close, 0);
        const sma = sum / slice.length;

        return {
            time: Math.floor(candle.openTime / 1000),
            value: sma,
        };
    }).filter(item => item !== null);

    // Обновляем данные на графике
    candleSeries.setData(candleData);
    volumeSeries.setData(volumeData);
    sma20Series.setData(sma20Data);
    sma50Series.setData(sma50Data);

    // Автоматическое масштабирование
    chart.timeScale().fitContent();
}

function updateIndicators(indicators) {
    // RSI
    const rsiValue = document.getElementById('rsiValue');
    const rsiStatus = document.getElementById('rsiStatus');
    rsiValue.textContent = indicators.RSI ? indicators.RSI.toFixed(2) : '-';

    if (indicators.RSI > 70) {
        rsiStatus.textContent = 'Перекупленность';
        rsiStatus.className = 'indicator-status status-bearish';
    } else if (indicators.RSI < 30) {
        rsiStatus.textContent = 'Перепроданность';
        rsiStatus.className = 'indicator-status status-bullish';
    } else {
        rsiStatus.textContent = 'Нейтрально';
        rsiStatus.className = 'indicator-status status-neutral';
    }

    // SMA20
    const sma20Value = document.getElementById('sma20Value');
    const sma20Status = document.getElementById('sma20Status');
    sma20Value.textContent = indicators.SMA20 ? indicators.SMA20.toFixed(2) : '-';

    // SMA50
    const sma50Value = document.getElementById('sma50Value');
    const sma50Status = document.getElementById('sma50Status');
    sma50Value.textContent = indicators.SMA50 ? indicators.SMA50.toFixed(2) : '-';

    // Сравниваем SMA20 и SMA50 для тренда
    if (indicators.SMA20 && indicators.SMA50) {
        if (indicators.SMA20 > indicators.SMA50) {
            sma20Status.textContent = '📈 Выше SMA50';
            sma20Status.className = 'indicator-status status-bullish';
            sma50Status.textContent = '📉 Ниже SMA20';
            sma50Status.className = 'indicator-status status-bearish';
        } else {
            sma20Status.textContent = '📉 Ниже SMA50';
            sma20Status.className = 'indicator-status status-bearish';
            sma50Status.textContent = '📈 Выше SMA20';
            sma50Status.className = 'indicator-status status-bullish';
        }
    }

    // MACD
    const macdValue = document.getElementById('macdValue');
    const macdStatus = document.getElementById('macdStatus');
    macdValue.textContent = indicators.MACD ? indicators.MACD.toFixed(4) : '-';

    if (indicators.MACD > 0) {
        macdStatus.textContent = 'Бычий';
        macdStatus.className = 'indicator-status status-bullish';
    } else {
        macdStatus.textContent = 'Медвежий';
        macdStatus.className = 'indicator-status status-bearish';
    }
}

function updateAnalysisText(data) {
    const analysisText = document.getElementById('analysisText');
    const indicators = data.indicators;

    let analysis = '';

    // Анализ тренда по SMA
    if (indicators.SMA20 && indicators.SMA50) {
        if (indicators.SMA20 > indicators.SMA50) {
            analysis += '📈 <strong>Восходящий тренд</strong> - SMA20 выше SMA50<br>';
        } else {
            analysis += '📉 <strong>Нисходящий тренд</strong> - SMA20 ниже SMA50<br>';
        }
    }

    // Анализ RSI
    if (indicators.RSI > 70) {
        analysis += '⚠️ <strong>Перекупленность</strong> - RSI выше 70. Возможна коррекция<br>';
    } else if (indicators.RSI < 30) {
        analysis += '🔄 <strong>Перепроданность</strong> - RSI ниже 30. Возможен отскок<br>';
    } else {
        analysis += '⚖️ <strong>Нейтральная зона</strong> - RSI в диапазоне 30-70<br>';
    }

    // Анализ MACD
    if (indicators.MACD > 0) {
        analysis += '🐂 <strong>Бычий сигнал MACD</strong> - гистограмма выше нуля<br>';
    } else {
        analysis += '🐻 <strong>Медвежий сигнал MACD</strong> - гистограмма ниже нуля<br>';
    }

    // Общие рекомендации
    analysis += '<br><strong>Рекомендации:</strong><br>';

    if (indicators.RSI > 70 && indicators.MACD > 0) {
        analysis += '• Рассмотрите возможность фиксации прибыли<br>';
        analysis += '• Осторожно с новыми покупками<br>';
    } else if (indicators.RSI < 30 && indicators.MACD < 0) {
        analysis += '• Возможность для покупки по низким ценам<br>';
        analysis += '• Установите стоп-лоссы<br>';
    } else if (indicators.SMA20 > indicators.SMA50 && indicators.MACD > 0) {
        analysis += '• Сильный восходящий тренд<br>';
        analysis += '• Рассмотрите покупки на коррекциях<br>';
    } else {
        analysis += '• Рынок в неопределенности<br>';
        analysis += '• Дождитесь четких сигналов<br>';
    }

    analysisText.innerHTML = analysis;
}

function showError(message) {
    const analysisText = document.getElementById('analysisText');
    analysisText.innerHTML = `<div style="color: var(--error-color);">❌ ${message}</div>`;
}