document.addEventListener('DOMContentLoaded', () => {
	// Установка минимальной даты и управление часами/минутами
	const today = new Date();
	const dateInput = document.getElementById('booking-date');
	const hourSelect = document.getElementById('booking-hour');
	const minuteSelect = document.getElementById('booking-minute');

	// Установка минимальной даты (сегодня)
	const todayStr = today.toISOString().split('T')[0];
	dateInput.setAttribute('min', todayStr);
	dateInput.value = todayStr; // Устанавливаем текущую дату по умолчанию

	const pathSegments = window.location.pathname.split('/').filter(segment => segment);
	const partnerId = pathSegments[pathSegments.length - 1];

	if (partnerId) {
		fetch(`/api/partner/${partnerId}`)
			.then(response => response.json())
			.then(data => {
				console.log('Данные партнера:', data);
				document.getElementById('partner-name').textContent = data.name;
				document.getElementById('partner-address').textContent = data.address;
				document.getElementById('partner-phone').textContent = data.phone;
				document.getElementById('partner-description').textContent = data.description || 'Описание отсутствует';

				const logoImg = document.getElementById('partner-logo-img');
				if (data.logoPath) {
					logoImg.src = `/${data.logoPath}`;
				} else {
					const logoContainer = document.getElementById('partner-logo');
					logoContainer.classList.add('placeholder');
					logoContainer.textContent = 'Логотип отсутствует';
					logoImg.style.display = 'none';
				}
			})
			.catch(error => {
				console.error('Ошибка загрузки данных партнера:', error);
			});
	}

	// Обработка отправки формы записи
	const bookingForm = document.getElementById('submit-booking');
	if (bookingForm) {
		bookingForm.addEventListener('click', () => {
			const bookingDate = document.getElementById('booking-date').value;
			const bookingHour = document.getElementById('booking-hour').value;
			const bookingMinute = document.getElementById('booking-minute').value;

			// Проверка на заполненность полей
			if (!bookingDate || !bookingHour || !bookingMinute) {
				showBookingMessage('Пожалуйста, выберите дату и время.', 'error');
				return;
			}

			// Формируем время в формате HH:MM
			const bookingTime = `${bookingHour}:${bookingMinute}`;

			const bookingData = {
				partner_id: parseInt(partnerId),
				booking_date: bookingDate,
				booking_time: bookingTime,
				status: 'pending'
			};

			fetch('/api/bookings', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify(bookingData),
			})
				.then(response => response.json())
				.then(data => {
					if (data.error) {
						showBookingMessage(data.error, 'error');
					} else {
						showBookingMessage('Запись успешно создана! Ожидайте подтверждения.', 'success');
						loadBookings();
					}
				})
				.catch(error => {
					console.error('Ошибка отправки записи:', error);
					showBookingMessage('Не удалось создать запись. Попробуйте снова.', 'error');
				});
		});
	}
	// Инициализация при загрузке страницы
	updateAvailableTimes();

	// Обновление при изменении даты
	dateInput.addEventListener('change', updateAvailableTimes);

	// Обновление минут при изменении часов
	hourSelect.addEventListener('change', () => {
		const selectedDate = new Date(dateInput.value);
		const isToday = selectedDate.toDateString() === today.toDateString();
		const selectedHour = parseInt(hourSelect.value);

		if (isToday && selectedHour === today.getHours()) {
			const currentMinute = today.getMinutes();
			const next15Minute = Math.ceil(currentMinute / 15) * 15;
			populateMinutes(next15Minute >= 60 ? 0 : next15Minute);
		} else {
			populateMinutes(0);
		}
	});
});

// Функция для отображения сообщений о бронировании
function showBookingMessage(message, type) {
	const messageElement = document.getElementById('booking-message');
	messageElement.textContent = message;
	messageElement.className = type === 'error' ? 'error-message' : 'success-message';
	messageElement.style.display = 'block';
	setTimeout(() => {
		messageElement.style.display = 'none';
	}, 5000);
}

// Загрузка записей для дашборда
function loadBookings() {
	fetch('/api/bookings')
		.then(response => response.json())
		.then(bookings => {
			const bookingsList = document.getElementById('bookings-list');
			bookingsList.innerHTML = ' ';

			if (bookings.length === 0) {
				bookingsList.innerHTML = '<p>Записей нет.</p>';
				return;
			}

			bookings.forEach(booking => {
				const bookingCard = document.createElement('div');
				bookingCard.className = 'booking-card';
				bookingCard.innerHTML = `
                    <div class="booking-info">
                        <p><strong>ID пользователя:</strong> ${booking.user_id}</p>
                        <p><strong>Дата:</strong> ${booking.booking_date}</p>
                        <p><strong>Время:</strong> ${booking.booking_time}</p>
                        <p><strong>Статус:</strong> ${booking.status}</p>
                    </div>
                    <div class="booking-actions">
                        <button onclick="updateBookingStatus(${booking.id}, 'confirmed')">Подтвердить</button>
                        <button onclick="updateBookingStatus(${booking.id}, 'canceled')">Отменить</button>
                    </div>
                `;
				bookingsList.appendChild(bookingCard);
			});
		})
		.catch(error => {
			console.error('Ошибка загрузки записей:', error);
			document.getElementById('bookings-list').innerHTML = '<p>Не удалось загрузить записи.</p>';
		});
}

// Обновление статуса записи
function updateBookingStatus(bookingId, status) {
	fetch(`/api/bookings/${bookingId}`, {
		method: 'PUT',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({ status }),
	})
		.then(response => response.json())
		.then(data => {
			if (data.error) {
				alert(data.error);
			} else {
				alert(`Запись успешно ${status === 'confirmed' ? 'подтверждена' : 'отменена'}!`);
				loadBookings();
			}
		})
		.catch(error => {
			console.error('Ошибка обновления статуса:', error);
			alert('Не удалось обновить статус записи.');
		});
}
// Функция для заполнения часов
function populateHours(minHour = 0) {
	hourSelect.innerHTML = '<option value="" disabled selected>Часы</option>';
	for (let h = minHour; h < 24; h++) {
		const hour = h.toString().padStart(2, '0');
		const option = document.createElement('option');
		option.value = hour;
		option.textContent = hour;
		hourSelect.appendChild(option);
	}
}

// Функция для заполнения минут
function populateMinutes(minMinute = 0) {
	minuteSelect.innerHTML = '<option value="" disabled selected>Минуты</option>';
	const minuteOptions = [0, 15, 30, 45].filter(min => min >= minMinute);
	minuteOptions.forEach(min => {
		const minute = min.toString().padStart(2, '0');
		const option = document.createElement('option');
		option.value = minute;
		option.textContent = minute;
		minuteSelect.appendChild(option);
	});
}

// Обновление доступного времени в зависимости от выбранной даты
function updateAvailableTimes() {
	const selectedDate = new Date(dateInput.value);
	const isToday = selectedDate.toDateString() === today.toDateString();
	hourSelect.innerHTML = '<option value="" disabled selected>Часы</option>';
	minuteSelect.innerHTML = '<option value="" disabled selected>Минуты</option>';

	if (isToday) {
		const currentHour = today.getHours();
		const currentMinute = today.getMinutes();
		const next15Minute = Math.ceil(currentMinute / 15) * 15; // Округляем до следующего кратного 15

		// Если текущее время, например, 14:37, минимальное доступное время — 14:45
		if (next15Minute >= 60) {
			populateHours(currentHour + 1);
			populateMinutes(0);
		} else {
			populateHours(currentHour);
			populateMinutes(next15Minute);
		}
	} else {
		// Для будущих дней доступны все часы и минуты
		populateHours(0);
		populateMinutes(0);
	}
}