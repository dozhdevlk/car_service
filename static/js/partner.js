document.addEventListener('DOMContentLoaded', () => {
	const pathSegments = window.location.pathname.split('/').filter(segment => segment);
	const partnerId = pathSegments[pathSegments.length - 1];
	const dateInput = document.getElementById('booking-date');
	const timeSelect = document.getElementById('booking-time'); // Один список для времени

	// Устанавливаем минимальную дату на сегодня
	const today = new Date();
	const todayStr = today.toISOString().split('T')[0];
	dateInput.setAttribute('min', todayStr);
	dateInput.value = todayStr; // Устанавливаем текущую дату по умолчанию

	let workingHours = {}; // Храним рабочие часы партнера

	function fetchPartnerAnnouncements(partnerId) {
		fetch(`/api/announcements/${partnerId}`)
			.then(response => response.json())
			.then(announcements => {
				renderAnnouncements(announcements); // Рендерим услуги
			})
			.catch(error => {
				console.error('Ошибка загрузки услуг партнера:', error);
				document.getElementById('services-list').innerHTML = '<p>Не удалось загрузить услуги.</p>';
			});
	}

	function fetchPartnerOfferings(partnerId) {
		fetch(`/api/partner_offerings?partner_id=${partnerId}`)
			.then(response => response.json())
			.then(offerings => {
				renderServices(offerings); // Рендерим услуги
			})
			.catch(error => {
				console.error('Ошибка загрузки услуг партнера:', error);
				document.getElementById('services-list').innerHTML = '<p>Не удалось загрузить услуги.</p>';
			});
	}

	function renderAnnouncements(announcements) {

		const announcemensList = document.getElementById('announcemens-list');
		announcemensList.innerHTML = '';

		announcemensList.forEach(announcement => {
			const announcementCard = document.createElement('div');
			announcementCard.className = 'announcement-card';
			announcementCard.innerHTML = `
			<img src=/${announcement.image_url} alt="Объявление ${announcement.id}">
			<h3>${announcement.title}</h3>
			<p>${announcement.text}</p>
			`;
			announcemensList.appendChild(announcementCard);
		});
	}
	// Функция для рендеринга списка услуг
	function renderServices(offerings) {
		const searchTerm = document.getElementById('service-search').value.toLowerCase();
		const servicesList = document.getElementById('services-list');
		servicesList.innerHTML = '';

		// Фильтрация по поисковому запросу
		const filteredServices = offerings.filter(service =>
			service.name.toLowerCase().includes(searchTerm)
		);

		// Отображение услуг
		filteredServices.forEach(service => {
			const serviceCard = document.createElement('div');
			serviceCard.className = 'service-card';
			serviceCard.innerHTML = `
                <div class="service-details">
                    <h3>${service.name}</h3>
                    <p>Цена: ${service.price.toLocaleString('ru-RU')} руб.</p>
                </div>
            `;
			servicesList.appendChild(serviceCard);
		});
	}

	// Обработчик поиска
	document.getElementById('service-search').addEventListener('input', () => {
		fetchPartnerOfferings(partnerId); // Перезагружаем список с фильтрацией
	});

	// Инициализация загрузки услуг при загрузке страницы
	fetchPartnerOfferings(partnerId);

	// Загружаем данные партнера
	fetch(`/api/partner/${partnerId}`)
		.then(response => response.json())
		.then(data => {
			workingHours = data.working_hours;
			// После загрузки данных обновляем доступные часы
			updateAvailableTimes();
		})
		.catch(error => {
			console.error('Ошибка загрузки данных партнера:', error);
		});


	// Обновление доступного времени в зависимости от выбранной даты
	function updateAvailableTimes() {
		const selectedDate = dateInput.value;

		fetch('/api/available-times', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify({
				partner_id: parseInt(partnerId),
				booking_date: selectedDate,
			})
		})
			.then(response => response.json())
			.then(data => {
				timeSelect.innerHTML = ''; // Очищаем текущие данные в списке
				const option = document.createElement('option');
				option.value = '';
				option.disabled = true;
				option.selected = true;
				option.textContent = 'Выберите время';
				timeSelect.appendChild(option);

				// Добавляем доступные временные промежутки
				data.forEach(timeSlot => {
					const option = document.createElement('option');
					option.value = timeSlot;
					option.textContent = timeSlot;
					timeSelect.appendChild(option);
				});
			})
			.catch(error => {
				console.error('Ошибка получения доступных слотов:', error);
			});
	}


	// Инициализация при загрузке страницы
	updateAvailableTimes();

	// Обновление при изменении даты
	dateInput.addEventListener('change', updateAvailableTimes);


	// Проверка роли пользователя и владения сервисом
	function checkUserRoleAndOwnership() {
		// Сначала получаем данные текущего пользователя
		fetch('/api/user', {
			method: 'GET',
			credentials: 'include'
		})
			.then(response => {
				if (response.status === 401) {
					throw new Error('Unauthorized');
				}
				return response.json();
			})
			.then(user => {
				const userId = user.ID;
				const userRole = user.role;

				// Получаем данные партнёра
				return fetch(`/api/partner/${partnerId}`)
					.then(response => response.json())
					.then(partner => {
						// Проверяем, является ли пользователь владельцем или администратором
						const isOwner = partner.owner_id === userId;
						const isAdmin = userRole === 'admin';

						// Если пользователь — владелец или администратор, показываем кнопку "Управление"
						if (isOwner || isAdmin) {
							const manageBtn = document.getElementById('manage-btn');
							manageBtn.href = `/dashboard/${partnerId}`;
							manageBtn.style.display = 'inline-block';
						}
					});
			})
			.catch(error => {
				console.error('Ошибка проверки роли или владения:', error);
				// Если пользователь не авторизован или произошла ошибка, кнопка остаётся скрытой
			});
	}

	// Загрузка данных партнера
	if (partnerId) {
		fetch(`/api/partner/${partnerId}`)
			.then(response => response.json())
			.then(data => {
				document.getElementById('partner-name').textContent = data.name;
				document.getElementById('partner-address').textContent = data.address;
				document.getElementById('partner-phone').textContent = data.phone;
				document.getElementById('partner-description').textContent = data.description || 'Описание отсутствует';

				// Отображение рабочих часов
				if (data.working_hours) {
					const workingHours = formatWorkingHours(data.working_hours);
					document.getElementById('partner-working-hours').textContent = workingHours;
				} else {
					document.getElementById('partner-working-hours').textContent = 'Не указаны.';
				}

				const logoImg = document.getElementById('partner-logo-img');
				if (data.logoPath) {
					logoImg.src = `/${data.logoPath}`;
				} else {
					const logoContainer = document.getElementById('partner-logo');
					logoContainer.classList.add('placeholder');
					logoContainer.textContent = 'Логотип отсутствует';
					logoImg.style.display = 'none';
				}

				checkUserRoleAndOwnership();
			})
			.catch(error => {
				console.error('Ошибка загрузки данных партнера:', error);
			});
	}

	// Функция для форматирования рабочих часов
	function formatWorkingHours(workingHours) {
		const daysOfWeek = ['Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота', 'Воскресенье'];
		const daysOfWeekEn = ['mon', 'tue', 'wed', 'thu', 'fri', 'sat', 'sun'];
		let hoursString = '';

		daysOfWeek.forEach((day, index) => {
			const dayKey = daysOfWeekEn[index].toLowerCase(); // "mon", "tue", "wed", ...
			const hours = workingHours[dayKey];

			if (hours.from != hours.to) {
				hoursString += `${day}: ${hours.from} - ${hours.to}\n`;
			} else {
				hoursString += `${day}: Выходной\n`;
			}
		});

		return hoursString;
	}

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

	// Обработка отправки формы записи
	const bookingForm = document.getElementById('submit-booking');
	if (bookingForm) {
		bookingForm.addEventListener('click', () => {
			const bookingDate = document.getElementById('booking-date').value;
			const bookingTime = document.getElementById('booking-time').value;

			// Проверка на заполненность полей
			if (!bookingDate || !bookingTime) {
				showBookingMessage('Пожалуйста, выберите дату и время.', 'error');
				return;
			}



			// Дополнительная проверка: выбранное время не в прошлом
			const selectedDateTime = new Date(bookingDate);
			selectedDateTime.setHours(parseInt(bookingTime.split(":")[0]), parseInt(bookingTime.split(":")[1]));
			const now = new Date();
			if (selectedDateTime <= now) {
				showBookingMessage('Нельзя записаться на прошедшее время.', 'error');
				return;
			}

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
						document.getElementById('booking-form').reset();
						loadBookings();
					}
				})
				.catch(error => {
					console.error('Ошибка отправки записи:', error);
					showBookingMessage('Не удалось создать запись. Попробуйте снова.', 'error');
				});
		});
	}

	// Функции для отображения сообщений
	function showBookingMessage(message, type) {
		const messageElement = document.getElementById('booking-message');
		messageElement.textContent = message;
		messageElement.className = type === 'error' ? 'error-message' : 'success-message';
		messageElement.style.display = 'block';
		setTimeout(() => { messageElement.style.display = 'none'; }, 5000);
	}

	function showServiceMessage(message, type) {
		const messageElement = document.getElementById('service-message');
		messageElement.textContent = message;
		messageElement.className = type === 'error' ? 'error-message' : 'success-message';
		messageElement.style.display = 'block';
		setTimeout(() => { messageElement.style.display = 'none'; }, 5000);
	}



});

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