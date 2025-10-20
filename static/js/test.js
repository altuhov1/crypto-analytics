function addDragToPan(canvas, chart, data) {
    let isDragging = false;
    let startX = 0;
    let startVisibleStart = visibleStart;
    let startVisibleEnd = visibleEnd;

    canvas.style.cursor = 'grab';

    canvas.addEventListener('mousedown', function (e) {
        isDragging = true;
        startX = e.clientX;
        startVisibleStart = visibleStart;
        startVisibleEnd = visibleEnd;
        canvas.style.cursor = 'grabbing';
        e.preventDefault();
    });

    canvas.addEventListener('mousemove', function (e) {
        if (!isDragging) return;

        const deltaX = e.clientX - startX;
        const totalVisiblePoints = startVisibleEnd - startVisibleStart;

        // Вычисляем смещение в единицах данных (точках)
        const movePoints = Math.round((deltaX / canvas.offsetWidth) * totalVisiblePoints);

        // Обновляем видимую область
        visibleStart = startVisibleStart - movePoints;
        visibleEnd = startVisibleEnd - movePoints;

        // Ограничиваем границы
        const maxIndex = data.labels.length - 1;
        if (visibleStart < 0) {
            visibleStart = 0;
            visibleEnd = totalVisiblePoints;
        }
        if (visibleEnd > maxIndex) {
            visibleEnd = maxIndex;
            visibleStart = maxIndex - totalVisiblePoints;
        }

        // Обновляем график
        updateVisibleRange(chart, data, visibleStart, visibleEnd);
    });

    canvas.addEventListener('mouseup', function () {
        isDragging = false;
        canvas.style.cursor = 'grab';
    });

    canvas.addEventListener('mouseleave', function () {
        isDragging = false;
        canvas.style.cursor = 'grab';
    });

    // Обработчик колесика мыши для зума
    canvas.addEventListener('wheel', function (e) {
        e.preventDefault();

        const zoomFactor = e.deltaY > 0 ? 1.2 : 0.8;
        const centerIndex = Math.round((visibleStart + visibleEnd) / 2);
        const currentRange = visibleEnd - visibleStart;
        const newRange = Math.round(currentRange * zoomFactor);

        // Ограничиваем зум
        const minRange = 5;
        const maxRange = data.labels.length;

        if (newRange >= minRange && newRange <= maxRange) {
            visibleStart = Math.max(0, centerIndex - Math.floor(newRange / 2));
            visibleEnd = Math.min(data.labels.length - 1, visibleStart + newRange);

            updateVisibleRange(chart, data, visibleStart, visibleEnd);
        }
    });
}