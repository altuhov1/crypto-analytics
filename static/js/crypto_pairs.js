let allPairs = [];
let selectedPair = null;
let searchTimeout = null;

// Элементы DOM
const searchBox = document.getElementById('searchBox');
const searchResults = document.getElementById('searchResults');
const resultsCount = document.getElementById('resultsCount');
const selectedPairSection = document.getElementById('selectedPairSection');
const selectedPairDisplay = document.getElementById('selectedPairDisplay');
const sendBtn = document.getElementById('sendBtn');
const clearBtn = document.getElementById('clearBtn');
const statusMessage = document.getElementById('statusMessage');

// Загрузка всех пар при старте
document.addEventListener('DOMContentLoaded', async function () {
    // Инициализация бургер-меню
    initBurgerMenu();

    await loadAllPairs();
    setupEventListeners();
});

// Функционал бургер-меню
function initBurgerMenu() {
    const burgerMenu = document.getElementById('burger-menu');
    const mainNav = document.getElementById('main-nav');
    const navOverlay = document.getElementById('nav-overlay');

    if (burgerMenu && mainNav && navOverlay) {
        burgerMenu.addEventListener('click', function (e) {
            e.stopPropagation();
            this.classList.toggle('active');
            mainNav.classList.toggle('active');
            navOverlay.classList.toggle('active');
        });

        navOverlay.addEventListener('click', function () {
            burgerMenu.classList.remove('active');
            mainNav.classList.remove('active');
            this.classList.remove('active');
        });

        const navLinks = mainNav.querySelectorAll('a');
        navLinks.forEach(link => {
            link.addEventListener('click', () => {
                burgerMenu.classList.remove('active');
                mainNav.classList.remove('active');
                navOverlay.classList.remove('active');
            });
        });

        window.addEventListener('resize', () => {
            if (window.innerWidth > 768) {
                burgerMenu.classList.remove('active');
                mainNav.classList.remove('active');
                navOverlay.classList.remove('active');
            }
        });
    }
}

// Загрузка всех пар с сервера
async function loadAllPairs() {
    try {
        showLoading('Загрузка списка пар...');

        const response = await fetch('/api/all-pairs');
        const data = await response.json();

        if (data.success) {
            allPairs = data.pairs;
            showInitialState();
            console.log(`Loaded ${allPairs.length} pairs`);
        } else {
            showError('Ошибка при загрузке пар: ' + data.error);
        }
    } catch (error) {
        console.error('Load pairs error:', error);
        showError('Ошибка при загрузке списка пар');
    }
}

function setupEventListeners() {
    // Поиск при вводе
    searchBox.addEventListener('input', function () {
        const query = this.value.trim().toUpperCase();

        if (searchTimeout) {
            clearTimeout(searchTimeout);
        }

        searchTimeout = setTimeout(() => {
            searchPairs(query);
        }, 100);
    });

    // Популярные пары
    document.querySelectorAll('.popular-pair').forEach(pair => {
        pair.addEventListener('click', function () {
            selectPair(this.dataset.pair);
            searchBox.value = '';
            showInitialState();
        });
    });

    // Очистка выбора
    clearBtn.addEventListener('click', clearSelection);

    // Отправка пары
    sendBtn.addEventListener('click', sendSelectedPair);
}

// Поиск пар (клиентский)
function searchPairs(query) {
    if (!query) {
        showInitialState();
        return;
    }

    const filteredPairs = allPairs.filter(pair =>
        pair.includes(query)
    );

    displayPairs(filteredPairs, query);
}

// Отображение состояния загрузки
function showLoading(message) {
    searchResults.innerHTML = `
                <div class="no-results">
                    <div class="loading" style="margin: 0 auto 15px;"></div>
                    <p>${message}</p>
                </div>
            `;
    resultsCount.textContent = message;
}

// Отображение начального состояния
function showInitialState() {
    const initialPairs = allPairs.slice(0, 20);
    displayPairs(initialPairs, '');
    resultsCount.textContent = `Всего пар: ${allPairs.length}. Введите текст для поиска`;
}

// Отображение ошибки
function showError(message) {
    searchResults.innerHTML = `
                <div class="no-results">
                    <i class="fas fa-exclamation-triangle" style="font-size: 2rem; margin-bottom: 10px; color: var(--error-color);"></i>
                    <p>${message}</p>
                </div>
            `;
    resultsCount.textContent = 'Ошибка загрузки';
}

// Отображение найденных пар
function displayPairs(pairs, query) {
    if (pairs.length === 0) {
        searchResults.innerHTML = `
            <div class="no-results">
                <i class="fas fa-search" style="font-size: 2rem; margin-bottom: 10px; opacity: 0.5;"></i>
                <p>Пар не найдено для "${query}"</p>
            </div>
        `;
        resultsCount.textContent = `Найдено пар: 0`;
        return;
    }

    const pairsHtml = pairs.map(pair => {
        const isSelected = selectedPair === pair;

        return `
    <div class="pair-item ${isSelected ? 'selected' : ''}" data-pair="${pair}">
        <span class="pair-symbol">${pair}</span>
        <button class="select-btn" ${isSelected ? 'disabled' : ''}>
            ${isSelected ? '✓ Выбрано' : 'Выбрать'}
        </button>
    </div>
    `;
    }).join('');

    searchResults.innerHTML = pairsHtml;
    resultsCount.textContent = `Найдено пар: ${pairs.length}${query ? ` для "${query}"` : ''}`;

    document.querySelectorAll('.pair-item').forEach(item => {
        item.addEventListener('click', function () {
            selectPair(this.dataset.pair);
        });
    });
}

// Выбор пары
function selectPair(pair) {
    selectedPair = pair;
    selectedPairDisplay.textContent = pair;
    selectedPairSection.style.display = 'block';
    updateSelectionInResults();
    selectedPairSection.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
}

// Обновление выбора в результатах
function updateSelectionInResults() {
    document.querySelectorAll('.pair-item').forEach(item => {
        if (item.dataset.pair === selectedPair) {
            item.classList.add('selected');
            item.querySelector('.select-btn').textContent = '✓ Выбрано';
            item.querySelector('.select-btn').disabled = true;
        } else {
            item.classList.remove('selected');
            item.querySelector('.select-btn').textContent = 'Выбрать';
            item.querySelector('.select-btn').disabled = false;
        }
    });
}

// Очистка выбора
function clearSelection() {
    selectedPair = null;
    selectedPairSection.style.display = 'none';
    statusMessage.style.display = 'none';

    document.querySelectorAll('.pair-item').forEach(item => {
        item.classList.remove('selected');
        item.querySelector('.select-btn').textContent = 'Выбрать';
        item.querySelector('.select-btn').disabled = false;
    });
}

// Отправка пары на бэкенд
async function sendSelectedPair() {
    if (!selectedPair) return;

    try {
        sendBtn.innerHTML = '<div class="loading"></div> Отправка...';
        sendBtn.disabled = true;
        statusMessage.style.display = 'none';

        const response = await fetch('/api/select-pair', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                pair: selectedPair
            })
        });

        const data = await response.json();

        if (data.success) {
            showStatus(data.message, 'success');
        } else {
            showStatus('Ошибка: ' + (data.error || data.message), 'error');
        }
    } catch (error) {
        console.error('Send error:', error);
        showStatus('Ошибка при отправке: ' + error.message, 'error');
    } finally {
        sendBtn.innerHTML = 'Отправить на анализ';
        sendBtn.disabled = false;
    }
}

// Показать статус
function showStatus(message, type) {
    statusMessage.textContent = message;
    statusMessage.className = `status-message status-${type}`;
    statusMessage.style.display = 'block';
}