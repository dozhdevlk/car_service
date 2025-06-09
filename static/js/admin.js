let allBookings = [];

document.addEventListener('DOMContentLoaded', () => {
	// Переход на вкладку "Сервисы" при клике на карточку "Новые сервисы"
	document.getElementById('new-services-card').addEventListener('click', () => {
		switchTab('services-content');
	});
	
	// Переход на вкладку "Все записи" при клике на карточку "Всего записей"
	document.getElementById('total-bookings-card').addEventListener('click', () => {
		switchTab('orders-content');
	});
	
	// Переход на вкладку "Пользователи" при клике на карточку "Всего пользователей"
	document.getElementById('total-users-card').addEventListener('click', () => {
		switchTab('users-content');
	});
	loadDashboard();



	// Переключение вкладок через боковую панель
	document.querySelectorAll('.tab-link').forEach(link => {
		link.addEventListener('click', (e) => {
			e.preventDefault();
			document.querySelectorAll('.tab-link').forEach(l => l.classList.remove('active'));
			document.querySelectorAll('.main-tab-content').forEach(c => {
				c.classList.remove('active');
				c.style.display = 'none';
			});

			link.classList.add('active');
			const tabId = link.getAttribute('data-tab');
			const tabContent = document.getElementById(tabId);
			tabContent.style.display = 'block';
			tabContent.classList.add('active');
			
			if (tabId === 'services-content') {
				renderServicesTable();
			}
			if (tabId === 'users-content') {
				renderUsersTable();
			}
			if (tabId === 'orders-content') {
				loadAllBookings();
			}
		});
	});

	const defaultTabLink = document.querySelector('.tab-link');
	if (defaultTabLink) {
		defaultTabLink.click();
	}

	document.getElementById('search-button').addEventListener('click', () => {
		const searchId = document.getElementById('search-id').value;
		const searchPhone = document.getElementById('search-phone').value;
		const selectedStatus = document.getElementById('status-filter').value;

		applyFilters(selectedStatus, searchId, searchPhone);
	});
});

function switchTab(tabId) {
	document.querySelectorAll('.main-tab-content').forEach(tab => {
		tab.style.display = 'none';
	});

	document.getElementById(tabId).style.display = 'block';
	document.querySelectorAll('.tab-link').forEach(link => {
		link.classList.remove('active');
	});
	document.querySelector(`.tab-link[data-tab="${tabId}"]`).classList.add('active');
}

function loadAllBookings() {
	fetch('/api/bookings')
		.then(response => {
			if (!response.ok) throw new Error('Ошибка при загрузке записей');
			return response.json();
		})
		.then(bookings => {
			allBookings = bookings;
			applyFilters('all');
		})
		.catch(error => {
			console.error('Ошибка загрузки записей:', error);
			document.getElementById('all-list').innerHTML = '<p>Не удалось загрузить записи.</p>';
		});
}

function applyFilters(status = '', searchId = '', searchPhone = '') {
	const filteredBookings = allBookings.filter(booking => {
		const isStatusMatch = ((status === 'all') || (booking.status === status));
		const isIdMatch = searchId ? booking.id.toString().includes(searchId) : true;
		const isPhoneMatch = searchPhone ? booking.user_phone.includes(searchPhone) : true;

		return isStatusMatch && isIdMatch && isPhoneMatch;
	});

	updateTabContent(filteredBookings);
}

function updateTabContent(bookings) {
	const list = document.getElementById('all-list');
	list.innerHTML = '';

	if (bookings.length === 0) {
		list.innerHTML = '<p>Записей не найдено.</p>';
	} else {
		bookings.forEach(booking => {
			const bookingCard = document.createElement('div');
			bookingCard.className = 'booking-card';
			bookingCard.innerHTML = `
                <div class="booking-info">
                    <p><strong>ID записи:</strong> ${booking.id}</p>
                    <p><strong>Название партнера:</strong> ${booking.partner_name}</p>
                    <p><strong>Телефон партнера:</strong> ${booking.partner_phone}</p>
                    <p><strong>Адрес партнера:</strong> ${booking.partner_address}</p>
                    <p><strong>Имя пользователя:</strong> ${booking.user_name}</p>
                    <p><strong>Телефон пользователя:</strong> ${booking.user_phone}</p>
                    <p><strong>Email пользователя:</strong> ${booking.user_email}</p>
                    <p><strong>Дата бронирования:</strong> ${booking.booking_date}</p>
                    <p><strong>Время бронирования:</strong> ${booking.booking_time}</p>
                    <p><strong>Статус:</strong> ${booking.status}</p>
                </div>
                <div class="booking-actions">
					${getActionButton(booking.status, booking.id)}
                </div>
            `;
			list.appendChild(bookingCard);
		});
	}
}


function getActionButton(status, bookingId) {
	switch (status) {
		case '⏳ Ожидает подтверждения':
			return `
                <button onclick="updateBookingStatus(${bookingId}, '✅ Подтверждена')">Подтвердить</button>
                <button onclick="updateBookingStatus(${bookingId}, '❌ Отменена')">Отменить</button>
            `;
		case '✅ Подтверждена':
			return `
                <button onclick="updateBookingStatus(${bookingId}, '🔧 В работе')">В работу</button>
            `;
		case '🔧 В работе':
			return `
                <button onclick="updateBookingStatus(${bookingId}, '🏁 Завершена')">Завершить</button>
            `;
		default:
			return '';
	}
}

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
				loadBookings(partnerId);
			}
		})
		.catch(error => {
			console.error('Ошибка обновления статуса:', error);
			alert('Не удалось обновить статус записи.');
		});
}

