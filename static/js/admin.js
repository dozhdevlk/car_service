document.addEventListener('DOMContentLoaded', () => {
	// Инициализация
	loadDashboard();
	renderServicesTable();
	renderUsersTable();

	// Переключение вкладок через боковую панель
	document.querySelectorAll('.tab-link').forEach(link => {
		link.addEventListener('click', (e) => {
			e.preventDefault();
			// Удаляем класс active у всех ссылок и скрываем только верхние контенты
			document.querySelectorAll('.tab-link').forEach(l => l.classList.remove('active'));
			document.querySelectorAll('.main-tab-content').forEach(c => {
				c.classList.remove('active');
				c.style.display = 'none';
			});

			// Активируем выбранную вкладку и её контент
			link.classList.add('active');
			const tabId = link.getAttribute('data-tab');
			const tabContent = document.getElementById(tabId);
			tabContent.style.display = 'block';
			tabContent.classList.add('active');
			console.log(`Активирована вкладка: ${tabId}, видимость: ${tabContent.style.display}`);

			// Загружаем данные только для нужной вкладки
			if (tabId === 'orders-content') {
				activateOrderTab('pending');
				loadBookings(); // Загружаем записи только при переключении на вкладку заказов
			}
		});
	});

	// Переключение внутренних вкладок заказов
	const tabs = document.querySelectorAll('.tab');
	tabs.forEach(tab => {
		tab.addEventListener('click', () => {
			tabs.forEach(t => t.classList.remove('active'));
			document.querySelectorAll('#orders-content .tab-content').forEach(content => {
				content.classList.remove('active');
				content.style.display = 'none';
			});

			tab.classList.add('active');
			const tabId = tab.getAttribute('data-tab');
			const tabContent = document.getElementById(tabId);
			tabContent.classList.add('active');
			tabContent.style.display = 'block';

			// Перезагружаем записи при переключении внутренних вкладок
			if (['pending', 'confirmed', 'canceled', 'working', 'end'].includes(tabId)) {
				loadBookings();
			}
		});
	});

	// Активируем первую вкладку (dashboard) при загрузке
	const defaultTabLink = document.querySelector('.tab-link');
	if (defaultTabLink) {
		defaultTabLink.click(); // Симулируем клик на первой вкладке
	}
});

function loadBookings() {
	fetch('/api/bookings')
		.then(response => {
			if (!response.ok) throw new Error('Ошибка при загрузке записей');
			return response.json();
		})
		.then(bookings => {
			const pendingList = document.getElementById('pending-list');
			const confirmedList = document.getElementById('confirmed-list');
			const canceledList = document.getElementById('canceled-list');
			const workingList = document.getElementById('working-list');
			const endList = document.getElementById('end-list');

			pendingList.innerHTML = '';
			confirmedList.innerHTML = '';
			canceledList.innerHTML = '';
			workingList.innerHTML = '';
			endList.innerHTML = '';

			const pendingBookings = bookings.filter(booking => booking.status === '⏳ Ожидает подтверждения');
			const confirmedBookings = bookings.filter(booking => booking.status === '✅ Подтверждена');
			const canceledBookings = bookings.filter(booking => booking.status === '❌ Отменена');
			const workingBookings = bookings.filter(booking => booking.status === '🔧 В работе')
			const endBookings = bookings.filter(booking => booking.status === '🏁 Завершена')


			if (pendingBookings.length === 0) {
				pendingList.innerHTML = '<p>Записей нет.</p>';
			} else {
				pendingBookings.forEach(booking => {
					const bookingCard = document.createElement('div');
					bookingCard.className = 'booking-card';
					bookingCard.innerHTML = `
						<div class="booking-info">
							<p><strong>ID записи:</strong> ${booking.id}</p>
							<p><strong>Название партнера:</strong> ${booking.partner_name}(${booking.partner_id})</p>
							<p><strong>Телефон партнера:</strong> ${booking.partner_phone}</p>
							<p><strong>Адрес партнера:</strong> ${booking.partner_address}</p>
							<p><strong>Имя пользователя:</strong> ${booking.user_name}(${booking.user_id})</p>
							<p><strong>Телефон пользователя:</strong> ${booking.user_phone}</p>
							<p><strong>Email пользователя:</strong> ${booking.user_email}</p>
							<p><strong>Дата бронирования:</strong> ${booking.booking_date}</p>
							<p><strong>Время бронирования:</strong> ${booking.booking_time}</p>
							<p><strong>Статус:</strong> ${booking.status}</p>
						</div>
                        <div class="booking-actions">
                            <button onclick="updateBookingStatus(${booking.id}, '✅ Подтверждена')">Подтвердить</button>
                            <button onclick="updateBookingStatus(${booking.id}, '❌ Отменена')">Отменить</button>
                        </div>
                    `;
					pendingList.appendChild(bookingCard);
				});
			}

			if (confirmedBookings.length === 0) {
				confirmedList.innerHTML = '<p>Записей нет.</p>';
			} else {
				confirmedBookings.forEach(booking => {
					const bookingCard = document.createElement('div');
					bookingCard.className = 'booking-card';
					bookingCard.innerHTML = `
					<div class="booking-info">
						<p><strong>ID записи:</strong> ${booking.id}</p>
						<p><strong>Название партнера:</strong> ${booking.partner_name}(${booking.partner_id})</p>
						<p><strong>Телефон партнера:</strong> ${booking.partner_phone}</p>
						<p><strong>Адрес партнера:</strong> ${booking.partner_address}</p>
						<p><strong>Имя пользователя:</strong> ${booking.user_name}(${booking.user_id})</p>
						<p><strong>Телефон пользователя:</strong> ${booking.user_phone}</p>
						<p><strong>Email пользователя:</strong> ${booking.user_email}</p>
						<p><strong>Дата бронирования:</strong> ${booking.booking_date}</p>
						<p><strong>Время бронирования:</strong> ${booking.booking_time}</p>
						<p><strong>Статус:</strong> ${booking.status}</p>
					</div>
					<div class="booking-actions">
                        <button onclick="updateBookingStatus(${booking.id}, '🔧 В работе')">Отправить в работу</button>
                        <button onclick="updateBookingStatus(${booking.id}, '❌ Отменена')">Отменить</button>
                    </div>
                    `;
					confirmedList.appendChild(bookingCard);
				});
			}

			if (canceledBookings.length === 0) {
				canceledList.innerHTML = '<p>Записей нет.</p>';
			} else {
				canceledBookings.forEach(booking => {
					const bookingCard = document.createElement('div');
					bookingCard.className = 'booking-card';
					bookingCard.innerHTML = `
					<div class="booking-info">
						<p><strong>ID записи:</strong> ${booking.id}</p>
						<p><strong>Название партнера:</strong> ${booking.partner_name}(${booking.partner_id})</p>
						<p><strong>Телефон партнера:</strong> ${booking.partner_phone}</p>
						<p><strong>Адрес партнера:</strong> ${booking.partner_address}</p>
						<p><strong>Имя пользователя:</strong> ${booking.user_name}(${booking.user_id})</p>
						<p><strong>Телефон пользователя:</strong> ${booking.user_phone}</p>
						<p><strong>Email пользователя:</strong> ${booking.user_email}</p>
						<p><strong>Дата бронирования:</strong> ${booking.booking_date}</p>
						<p><strong>Время бронирования:</strong> ${booking.booking_time}</p>
						<p><strong>Статус:</strong> ${booking.status}</p>
					</div>
                    `;
					canceledList.appendChild(bookingCard);
				});
			}
			if (workingBookings.length === 0) {
				workingList.innerHTML = '<p>Записей нет.</p>';
			} else {
				workingBookings.forEach(booking => {
					const bookingCard = document.createElement('div');
					bookingCard.className = 'booking-card';
					bookingCard.innerHTML = `
					<div class="booking-info">
						<p><strong>ID записи:</strong> ${booking.id}</p>
						<p><strong>Название партнера:</strong> ${booking.partner_name}(${booking.partner_id})</p>
						<p><strong>Телефон партнера:</strong> ${booking.partner_phone}</p>
						<p><strong>Адрес партнера:</strong> ${booking.partner_address}</p>
						<p><strong>Имя пользователя:</strong> ${booking.user_name}(${booking.user_id})</p>
						<p><strong>Телефон пользователя:</strong> ${booking.user_phone}</p>
						<p><strong>Email пользователя:</strong> ${booking.user_email}</p>
						<p><strong>Дата бронирования:</strong> ${booking.booking_date}</p>
						<p><strong>Время бронирования:</strong> ${booking.booking_time}</p>
						<p><strong>Статус:</strong> ${booking.status}</p>
					</div>
					<div class="booking-actions">
                        <button onclick="updateBookingStatus(${booking.id}, '🏁 Завершена')">Завершить</button>
                        <button onclick="updateBookingStatus(${booking.id}, '❌ Отменена')">Отменить</button>
                    </div>
			`;
					workingList.appendChild(bookingCard);
				});
			}
			if (endBookings.length === 0) {
				endList.innerHTML = '<p>Записей нет.</p>';
			} else {
				endBookings.forEach(booking => {
					const bookingCard = document.createElement('div');
					bookingCard.className = 'booking-card';
					bookingCard.innerHTML = `
					<div class="booking-info">
						<p><strong>ID записи:</strong> ${booking.id}</p>
						<p><strong>Название партнера:</strong> ${booking.partner_name}(${booking.partner_id})</p>
						<p><strong>Телефон партнера:</strong> ${booking.partner_phone}</p>
						<p><strong>Адрес партнера:</strong> ${booking.partner_address}</p>
						<p><strong>Имя пользователя:</strong> ${booking.user_name}(${booking.user_id})</p>
						<p><strong>Телефон пользователя:</strong> ${booking.user_phone}</p>
						<p><strong>Email пользователя:</strong> ${booking.user_email}</p>
						<p><strong>Дата бронирования:</strong> ${booking.booking_date}</p>
						<p><strong>Время бронирования:</strong> ${booking.booking_time}</p>
						<p><strong>Статус:</strong> ${booking.status}</p>
					</div>
			`;
					endList.appendChild(bookingCard);
				});
			}


		})
		.catch(error => {
			console.error('Ошибка загрузки записей:', error);
			document.getElementById('pending-list').innerHTML = '<p>Не удалось загрузить записи.</p>';
			document.getElementById('confirmed-list').innerHTML = '<p>Не удалось загрузить записи.</p>';
			document.getElementById('canceled-list').innerHTML = '<p>Не удалось загрузить записи.</p>';
		});
}

// Активация первой вкладки заказов
function activateOrderTab(defaultTab) {
	const tabs = document.querySelectorAll('.tab');
	const tabContents = document.querySelectorAll('#orders-content .tab-content'); // Ограничиваем только вложенными вкладками
	tabs.forEach(t => t.classList.remove('active'));
	tabContents.forEach(c => {
		c.classList.remove('active');
		c.style.display = 'none';
	});
	const activeTab = document.querySelector(`.tab[data-tab="${defaultTab}"]`);
	const activeContent = document.getElementById(defaultTab);
	if (activeTab && activeContent) {
		activeTab.classList.add('active');
		activeContent.classList.add('active');
		activeContent.style.display = 'block';
		console.log(`Активирована внутренняя вкладка: ${defaultTab}, видимость: ${activeContent.style.display}`);
	}
}

// Загрузка статистики
async function loadDashboard() {
	try {
		const response = await fetch('/api/admin/stats');
		if (!response.ok) throw new Error('Ошибка загрузки статистики');
		const data = await response.json();

		document.getElementById('pending-services').textContent = data.pending_services;
		document.getElementById('total-bookings').textContent = data.total_bookings;
		document.getElementById('total-users').textContent = data.total_users;
	} catch (error) {
		console.error('Error:', error);
	}
}

// Таблица сервисов
async function renderServicesTable() {
	try {
		const response = await fetch('/api/admin/services');
		if (!response.ok) throw new Error('Ошибка загрузки сервисов');
		const services = await response.json();

		const table = `
            <table class="moderation-table">
                <thead>
                    <tr>
                        <th>Название</th>
                        <th>Адрес</th>
                        <th>Телефон</th>
                        <th>Статус</th>
                        <th>Действия</th>
                    </tr>
                </thead>
                <tbody>
                    ${services.map(service => `
                        <tr>
                            <td>${service.name}</td>
                            <td>${service.address}</td>
                            <td>${service.phone}</td>
                            <td>${service.approved ? '✅ Одобрен' : '🔄 На проверке'}</td>
                            <td>
                                ${!service.approved ? `
                                    <button class="approve" onclick="approveService(${service.id})">Одобрить</button>
                                ` : `<button class="disapprove" onclick="disapproveService(${service.id})">На проверку</button>`}
                            </td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;
		document.getElementById('services-table').innerHTML = table;
	} catch (error) {
		console.error('Error:', error);
	}
}

// Одобрение сервиса
async function approveService(serviceId, flag) {
	try {
		const response = await fetch('/api/admin/approve-service', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ service_id: serviceId })
		});
		if (response.ok) {
			renderServicesTable();
		} else {
			alert('Ошибка при одобрении сервиса');
		}
	} catch (error) {
		console.error('Error:', error);
	}
}
async function disapproveService(serviceId) {
	try {
		const response = await fetch('/api/admin/disapprove-service', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ service_id: serviceId })
		});
		if (response.ok) {
			renderServicesTable();
		} else {
			alert('Ошибка при отправке сервиса на проверку');
		}
	} catch (error) {
		console.error('Error:', error);
	}
}

// Таблица пользователей
async function renderUsersTable() {
	try {
		const response = await fetch('/api/admin/users');
		if (!response.ok) throw new Error('Ошибка загрузки пользователей');
		const users = await response.json();

		const table = `
            <table class="moderation-table">
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Имя</th>
                        <th>Email</th>
						<th>Телефон</th>
                        <th>Роль</th>
						<th>Действия</th>
                    </tr>
                </thead>
                <tbody>
                    ${users.map(user => `
                        <tr>
                            <td>${user.id}</td>
                            <td>${user.name}</td>
                            <td>${user.email}</td>
							<td>${user.phone}</td>
                            <td>${user.role}</td>
							<td>
                                    <button class="deleteUsr" onclick="deleteUser(${user.id})">Удалить</button>
                            </td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;
		document.getElementById('users-table').innerHTML = table;
	} catch (error) {
		console.error('Error:', error);
	}
}
// удаление пользователя
async function deleteUser(userId) {
	try {
		const response = await fetch(`/api/admin/delete-user/${userId}`, {
			method: 'DELETE',
			headers: {
				'Content-Type': 'application/json'
			}
		});
		if (response.ok) {
			renderUsersTable();
		} else {
			alert('Ошибка при удалении пользователя');
		}
	} catch (error) {
		console.error('Error:', error);
	}
}


// Функция для обновления статуса записи (должна быть определена)
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
				alert(`Запись успешно обновлена`);
				// Перезагружаем записи, извлекая partnerId из URL
				const pathSegments = window.location.pathname.split('/').filter(segment => segment);
				const partnerId = pathSegments[pathSegments.length - 1];
				loadBookings(partnerId);
			}
		})
		.catch(error => {
			console.error('Ошибка обновления статуса:', error);
			alert('Не удалось обновить статус записи.');
		});
}

// Активация первой вкладки заказов
function activateOrderTab(defaultTab) {
	const tabs = document.querySelectorAll('.tab');
	const tabContents = document.querySelectorAll('.tab-content');
	tabs.forEach(t => t.classList.remove('active'));
	tabContents.forEach(c => {
		c.classList.remove('active');
		c.style.display = 'none';
	});
	const activeTab = document.querySelector(`.tab[data-tab="${defaultTab}"]`);
	const activeContent = document.getElementById(defaultTab);
	if (activeTab && activeContent) {
		activeTab.classList.add('active');
		activeContent.classList.add('active');
		activeContent.style.display = 'block';
	}
}