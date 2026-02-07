// Функционал бургер-меню
document.addEventListener('DOMContentLoaded', function () {
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
});

// Остальной существующий код для таблицы криптовалют
// Проверяем авторизацию при загрузке страницы
let isAuthenticated = false;
let favoritesSet = new Set();
let showFavoritesOnly = false;
let currentSearchValue = '';

// Функция поиска
function searchTable() {
    const input = document.getElementById('searchBox');
    const filter = input.value.toLowerCase();
    currentSearchValue = filter; // сохраняем текущий фильтр
    const rows = document.querySelectorAll('tbody tr');

    rows.forEach(row => {
        const name = row.querySelector('td:nth-child(2) strong').textContent.toLowerCase();
        const symbol = row.querySelector('td:nth-child(2) small').textContent.toLowerCase();

        if (name.includes(filter) || symbol.includes(filter)) {
            row.style.display = '';
        } else {
            row.style.display = 'none';
        }
    });

    updateRowNumbers();
}

// Инициализация поиска
document.getElementById('searchBox').addEventListener('input', function () {
    searchTable();
    // После поиска нужно заново применить фильтр "избранное", если он активен
    if (showFavoritesOnly) {
        filterTable();
    }
});

async function checkAuth() {
    try {
        const response = await fetch('/check-Sess-Id', {
            credentials: 'include'
        });
        const data = await response.json();

        if (data.authenticated) {
            isAuthenticated = true;
            enableHeartButtons();
            document.getElementById('favoritesToggle').disabled = false;
        } else {
            disableHeartButtons();
            document.getElementById('favoritesToggle').disabled = true;
        }
    } catch (error) {
        console.error('Auth check failed:', error);
        disableHeartButtons();
        document.getElementById('favoritesToggle').disabled = true;
    }
}

function enableHeartButtons() {
    const buttons = document.querySelectorAll('.heart-btn');
    buttons.forEach(btn => {
        btn.disabled = false;
        btn.title = 'Добавить в избранное';
    });
}

function disableHeartButtons() {
    const buttons = document.querySelectorAll('.heart-btn');
    buttons.forEach(btn => {
        btn.disabled = true;
        btn.title = 'Войдите, чтобы добавить в избранное';
    });
}

// Обработчик клика по сердечку
document.addEventListener('click', async function (e) {
    if (e.target.closest('.heart-btn')) {
        const button = e.target.closest('.heart-btn');

        if (!isAuthenticated) {
            alert('Пожалуйста, войдите в систему чтобы добавлять в избранное');
            return;
        }

        const coinId = button.dataset.coinId;
        const isLiked = button.classList.contains('liked');

        try {
            const response = await fetch('/api/changeFavoriteCoin', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({
                    coinId: coinId,
                    action: isLiked ? 'remove' : 'add'
                })
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`HTTP ${response.status}: ${errorText}`);
            }

            const result = await response.json();

            if (result.success) {
                if (isLiked) {
                    button.classList.remove('liked');
                    button.innerHTML = '<i class="far fa-heart"></i>';
                    favoritesSet.delete(coinId);
                } else {
                    button.classList.add('liked');
                    button.innerHTML = '<i class="fas fa-heart"></i>';
                    favoritesSet.add(coinId);
                }
                console.log('Успешно обновлено:', result.message);
                updateFavoritesDisplay();
                filterTable();
            } else {
                alert('Ошибка: ' + (result.message || 'Неизвестная ошибка'));
            }
        } catch (error) {
            console.error('Error:', error);
            alert('Ошибка при обновлении: ' + error.message);
        }
    }
});

// Загружаем избранные монеты при загрузке страницы
async function loadFavorites() {
    if (!isAuthenticated) return;

    try {
        const response = await fetch('/api/allFavoriteCoin', {
            credentials: 'include'
        });

        if (response.ok) {
            const result = await response.json();
            const favorites = result.data;
            favoritesSet = new Set(favorites);

            // Обновляем ВСЕ кнопки сердечек на основе favoritesSet
            const allHeartButtons = document.querySelectorAll('.heart-btn');
            allHeartButtons.forEach(button => {
                const coinId = button.dataset.coinId;
                if (favoritesSet.has(coinId)) {
                    button.classList.add('liked');
                    button.innerHTML = '<i class="fas fa-heart"></i>';
                } else {
                    button.classList.remove('liked');
                    button.innerHTML = '<i class="far fa-heart"></i>';
                }
            });

            updateFavoritesDisplay();
            filterTable();
        }
    } catch (error) {
        console.error('Error loading favorites:', error);
    }
}

// Обновляем отображение количества избранных
function updateFavoritesDisplay() {
    const countElement = document.querySelector('.favorites-count');
    const toggleBtn = document.getElementById('favoritesToggle');
    const count = favoritesSet.size;

    countElement.textContent = count;

    if (count === 0) {
        toggleBtn.disabled = true;
        showFavoritesOnly = false;
        toggleBtn.classList.remove('active');
        toggleBtn.innerHTML = `<i class="fas fa-heart"></i> Показать только избранное <span class="favorites-count">${count}</span>`;
    } else {
        toggleBtn.disabled = false;
    }
}

// Фильтрация таблицы (для избранного)
function filterTable() {
    const rows = document.querySelectorAll('tbody tr');

    rows.forEach(row => {
        const coinId = row.dataset.coinId;
        const isHiddenBySearch = row.style.display === 'none'; // уже скрыто поисковым фильтром

        if (showFavoritesOnly) {
            if (favoritesSet.has(coinId) && !isHiddenBySearch) {
                row.style.display = '';
            } else {
                row.style.display = 'none';
            }
        } else {
            // Если "показать все", возвращаем строки, которые соответствуют поиску
            if (currentSearchValue) {
                // Если есть поисковый фильтр — проверяем, соответствует ли строка ему
                const name = row.querySelector('td:nth-child(2) strong').textContent.toLowerCase();
                const symbol = row.querySelector('td:nth-child(2) small').textContent.toLowerCase();
                if (name.includes(currentSearchValue) || symbol.includes(currentSearchValue)) {
                    row.style.display = '';
                } else {
                    row.style.display = 'none';
                }
            } else {
                // Если поиска нет — показываем всё
                row.style.display = '';
            }
        }
    });

    updateRowNumbers();
}

// Обновляем нумерацию строк
function updateRowNumbers() {
    const visibleRows = Array.from(document.querySelectorAll('tbody tr')).filter(row => row.style.display !== 'none');
    visibleRows.forEach((row, index) => {
        const rankCell = row.querySelector('.crypto-rank');
        if (rankCell) {
            rankCell.textContent = index + 1;
        }
    });
}

// Обработчик кнопки переключения избранного
document.getElementById('favoritesToggle').addEventListener('click', function () {
    showFavoritesOnly = !showFavoritesOnly;
    const toggleBtn = this;

    if (showFavoritesOnly) {
        toggleBtn.classList.add('active');
        toggleBtn.innerHTML = `<i class="fas fa-heart"></i> Показать все монеты <span class="favorites-count">${favoritesSet.size}</span>`;
    } else {
        toggleBtn.classList.remove('active');
        toggleBtn.innerHTML = `<i class="fas fa-heart"></i> Показать только избранное <span class="favorites-count">${favoritesSet.size}</span>`;
    }

    filterTable();
});

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', async function () {
    await checkAuth();
    if (isAuthenticated) {
        await loadFavorites();
    }
});