let allBookings = [];

document.addEventListener('DOMContentLoaded', () => {
	// Инициализация
	loadDashboard();
	renderServicesTable();
	renderUsersTable();

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

			if (tabId === 'orders-content') {
				loadAllBookings();
			}
		});
	});


	// Активируем первую вкладку (dashboard) при загрузке
	const defaultTabLink = document.querySelector('.tab-link');
	if (defaultTabLink) {
		defaultTabLink.click(); // Симулируем клик на первой вкладке
	}

	// Поиск по ID записи или телефону владельца
	document.getElementById('search-button').addEventListener('click', () => {
		const searchId = document.getElementById('search-id').value;
		const searchPhone = document.getElementById('search-phone').value;
		const selectedStatus = document.getElementById('status-filter').value;

		applyFilters(selectedStatus, searchId, searchPhone); // Фильтрация записей по ID и телефону
	});
});


// Функция для загрузки всех записей и сохранения их в память
function loadAllBookings() {
	fetch('/api/bookings')
		.then(response => {
			if (!response.ok) throw new Error('Ошибка при загрузке записей');
			return response.json();
		})
		.then(bookings => {
			allBookings = bookings; // Сохраняем все записи в памяти
			applyFilters('all'); // Применяем фильтрацию сразу после загрузки данных
		})
		.catch(error => {
			console.error('Ошибка загрузки записей:', error);
			document.getElementById('all-list').innerHTML = '<p>Не удалось загрузить записи.</p>';
		});
}

function applyFilters(status = '', searchId = '', searchPhone = '') {
	const filteredBookings = allBookings.filter(booking => {
		const isStatusMatch =  booking.status === status; // Фильтрация по статусу
		const isIdMatch = searchId ? booking.id.toString().includes(searchId) : true; // Фильтрация по ID
		const isPhoneMatch = searchPhone ? booking.partner_phone.includes(searchPhone) : true; // Фильтрация по телефону

		return isStatusMatch && isIdMatch && isPhoneMatch; // Все условия должны совпасть
	});

	// Обновляем отображение записей
	updateTabContent(filteredBookings);
}

// Функция для обновления контента вкладки
function updateTabContent(bookings) {
	const list = document.getElementById(`${tabId}-list`);
	list.innerHTML = ''; // Очищаем текущий контент

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
			return ''; // Для других статусов кнопки не отображаются
	}
}
// // Функция для фильтрации записей по ID и телефону
// function filterRecords(searchId = '', searchPhone = '') {
// 	// Применяем фильтрацию ко всем вкладкам
// 	applyFilters('all', searchId, searchPhone); // Для вкладки "Все"
// 	applyFilters('pending', searchId, searchPhone); // Для вкладки "Ожидающие"
// 	applyFilters('confirmed', searchId, searchPhone); // Для вкладки "Подтвержденные"
// 	applyFilters('working', searchId, searchPhone); // Для вкладки "Записи в работе"
// 	applyFilters('end', searchId, searchPhone); // Для вкладки "Завершенные"
// 	applyFilters('canceled', searchId, searchPhone); // Для вкладки "Отмененные"
// }

// function loadBookings() {
// 	fetch('/api/bookings')
// 		.then(response => {
// 			if (!response.ok) throw new Error('Ошибка при загрузке записей');
// 			return response.json();
// 		})
// 		.then(bookings => {
// 			const pendingList = document.getElementById('pending-list');
// 			const confirmedList = document.getElementById('confirmed-list');
// 			const canceledList = document.getElementById('canceled-list');
// 			const workingList = document.getElementById('working-list');
// 			const endList = document.getElementById('end-list');

// 			pendingList.innerHTML = '';
// 			confirmedList.innerHTML = '';
// 			canceledList.innerHTML = '';
// 			workingList.innerHTML = '';
// 			endList.innerHTML = '';

// 			const pendingBookings = bookings.filter(booking => booking.status === '⏳ Ожидает подтверждения');
// 			const confirmedBookings = bookings.filter(booking => booking.status === '✅ Подтверждена');
// 			const canceledBookings = bookings.filter(booking => booking.status === '❌ Отменена');
// 			const workingBookings = bookings.filter(booking => booking.status === '🔧 В работе')
// 			const endBookings = bookings.filter(booking => booking.status === '🏁 Завершена')


// 			if (pendingBookings.length === 0) {
// 				pendingList.innerHTML = '<p>Записей нет.</p>';
// 			} else {
// 				pendingBookings.forEach(booking => {
// 					const bookingCard = document.createElement('div');
// 					bookingCard.className = 'booking-card';
// 					bookingCard.innerHTML = `
// 						<div class="booking-info">
// 							<p><strong>ID записи:</strong> ${booking.id}</p>
// 							<p><strong>Название партнера:</strong> ${booking.partner_name}(${booking.partner_id})</p>
// 							<p><strong>Телефон партнера:</strong> ${booking.partner_phone}</p>
// 							<p><strong>Адрес партнера:</strong> ${booking.partner_address}</p>
// 							<p><strong>Имя пользователя:</strong> ${booking.user_name}(${booking.user_id})</p>
// 							<p><strong>Телефон пользователя:</strong> ${booking.user_phone}</p>
// 							<p><strong>Email пользователя:</strong> ${booking.user_email}</p>
// 							<p><strong>Дата бронирования:</strong> ${booking.booking_date}</p>
// 							<p><strong>Время бронирования:</strong> ${booking.booking_time}</p>
// 							<p><strong>Статус:</strong> ${booking.status}</p>
// 						</div>
//                         <div class="booking-actions">
//                             <button onclick="updateBookingStatus(${booking.id}, '✅ Подтверждена')">Подтвердить</button>
//                             <button onclick="updateBookingStatus(${booking.id}, '❌ Отменена')">Отменить</button>
//                         </div>
//                     `;
// 					pendingList.appendChild(bookingCard);
// 				});
// 			}

// 			if (confirmedBookings.length === 0) {
// 				confirmedList.innerHTML = '<p>Записей нет.</p>';
// 			} else {
// 				confirmedBookings.forEach(booking => {
// 					const bookingCard = document.createElement('div');
// 					bookingCard.className = 'booking-card';
// 					bookingCard.innerHTML = `
// 					<div class="booking-info">
// 						<p><strong>ID записи:</strong> ${booking.id}</p>
// 						<p><strong>Название партнера:</strong> ${booking.partner_name}(${booking.partner_id})</p>
// 						<p><strong>Телефон партнера:</strong> ${booking.partner_phone}</p>
// 						<p><strong>Адрес партнера:</strong> ${booking.partner_address}</p>
// 						<p><strong>Имя пользователя:</strong> ${booking.user_name}(${booking.user_id})</p>
// 						<p><strong>Телефон пользователя:</strong> ${booking.user_phone}</p>
// 						<p><strong>Email пользователя:</strong> ${booking.user_email}</p>
// 						<p><strong>Дата бронирования:</strong> ${booking.booking_date}</p>
// 						<p><strong>Время бронирования:</strong> ${booking.booking_time}</p>
// 						<p><strong>Статус:</strong> ${booking.status}</p>
// 					</div>
// 					<div class="booking-actions">
//                         <button onclick="updateBookingStatus(${booking.id}, '🔧 В работе')">Отправить в работу</button>
//                         <button onclick="updateBookingStatus(${booking.id}, '❌ Отменена')">Отменить</button>
//                     </div>
//                     `;
// 					confirmedList.appendChild(bookingCard);
// 				});
// 			}

// 			if (canceledBookings.length === 0) {
// 				canceledList.innerHTML = '<p>Записей нет.</p>';
// 			} else {
// 				canceledBookings.forEach(booking => {
// 					const bookingCard = document.createElement('div');
// 					bookingCard.className = 'booking-card';
// 					bookingCard.innerHTML = `
// 					<div class="booking-info">
// 						<p><strong>ID записи:</strong> ${booking.id}</p>
// 						<p><strong>Название партнера:</strong> ${booking.partner_name}(${booking.partner_id})</p>
// 						<p><strong>Телефон партнера:</strong> ${booking.partner_phone}</p>
// 						<p><strong>Адрес партнера:</strong> ${booking.partner_address}</p>
// 						<p><strong>Имя пользователя:</strong> ${booking.user_name}(${booking.user_id})</p>
// 						<p><strong>Телефон пользователя:</strong> ${booking.user_phone}</p>
// 						<p><strong>Email пользователя:</strong> ${booking.user_email}</p>
// 						<p><strong>Дата бронирования:</strong> ${booking.booking_date}</p>
// 						<p><strong>Время бронирования:</strong> ${booking.booking_time}</p>
// 						<p><strong>Статус:</strong> ${booking.status}</p>
// 					</div>
//                     `;
// 					canceledList.appendChild(bookingCard);
// 				});
// 			}
// 			if (workingBookings.length === 0) {
// 				workingList.innerHTML = '<p>Записей нет.</p>';
// 			} else {
// 				workingBookings.forEach(booking => {
// 					const bookingCard = document.createElement('div');
// 					bookingCard.className = 'booking-card';
// 					bookingCard.innerHTML = `
// 					<div class="booking-info">
// 						<p><strong>ID записи:</strong> ${booking.id}</p>
// 						<p><strong>Название партнера:</strong> ${booking.partner_name}(${booking.partner_id})</p>
// 						<p><strong>Телефон партнера:</strong> ${booking.partner_phone}</p>
// 						<p><strong>Адрес партнера:</strong> ${booking.partner_address}</p>
// 						<p><strong>Имя пользователя:</strong> ${booking.user_name}(${booking.user_id})</p>
// 						<p><strong>Телефон пользователя:</strong> ${booking.user_phone}</p>
// 						<p><strong>Email пользователя:</strong> ${booking.user_email}</p>
// 						<p><strong>Дата бронирования:</strong> ${booking.booking_date}</p>
// 						<p><strong>Время бронирования:</strong> ${booking.booking_time}</p>
// 						<p><strong>Статус:</strong> ${booking.status}</p>
// 					</div>
// 					<div class="booking-actions">
//                         <button onclick="updateBookingStatus(${booking.id}, '🏁 Завершена')">Завершить</button>
//                         <button onclick="updateBookingStatus(${booking.id}, '❌ Отменена')">Отменить</button>
//                     </div>
// 			`;
// 					workingList.appendChild(bookingCard);
// 				});
// 			}
// 			if (endBookings.length === 0) {
// 				endList.innerHTML = '<p>Записей нет.</p>';
// 			} else {
// 				endBookings.forEach(booking => {
// 					const bookingCard = document.createElement('div');
// 					bookingCard.className = 'booking-card';
// 					bookingCard.innerHTML = `
// 					<div class="booking-info">
// 						<p><strong>ID записи:</strong> ${booking.id}</p>
// 						<p><strong>Название партнера:</strong> ${booking.partner_name}(${booking.partner_id})</p>
// 						<p><strong>Телефон партнера:</strong> ${booking.partner_phone}</p>
// 						<p><strong>Адрес партнера:</strong> ${booking.partner_address}</p>
// 						<p><strong>Имя пользователя:</strong> ${booking.user_name}(${booking.user_id})</p>
// 						<p><strong>Телефон пользователя:</strong> ${booking.user_phone}</p>
// 						<p><strong>Email пользователя:</strong> ${booking.user_email}</p>
// 						<p><strong>Дата бронирования:</strong> ${booking.booking_date}</p>
// 						<p><strong>Время бронирования:</strong> ${booking.booking_time}</p>
// 						<p><strong>Статус:</strong> ${booking.status}</p>
// 					</div>
// 			`;
// 					endList.appendChild(bookingCard);
// 				});
// 			}


// 		})
// 		.catch(error => {
// 			console.error('Ошибка загрузки записей:', error);
// 			document.getElementById('pending-list').innerHTML = '<p>Не удалось загрузить записи.</p>';
// 			document.getElementById('confirmed-list').innerHTML = '<p>Не удалось загрузить записи.</p>';
// 			document.getElementById('working-list').innerHTML = '<p>Не удалось загрузить записи.</p>';
// 			document.getElementById('end-list').innerHTML = '<p>Не удалось загрузить записи.</p>';
// 			document.getElementById('canceled-list').innerHTML = '<p>Не удалось загрузить записи.</p>';

// 		});
// }

// // Активация первой вкладки заказов
// function activateOrderTab(defaultTab) {
// 	const tabs = document.querySelectorAll('.tab');
// 	const tabContents = document.querySelectorAll('#orders-content .tab-content'); // Ограничиваем только вложенными вкладками
// 	tabs.forEach(t => t.classList.remove('active'));
// 	tabContents.forEach(c => {
// 		c.classList.remove('active');
// 		c.style.display = 'none';
// 	});
// 	const activeTab = document.querySelector(`.tab[data-tab="${defaultTab}"]`);
// 	const activeContent = document.getElementById(defaultTab);
// 	if (activeTab && activeContent) {
// 		activeTab.classList.add('active');
// 		activeContent.classList.add('active');
// 		activeContent.style.display = 'block';
// 		console.log(`Активирована внутренняя вкладка: ${defaultTab}, видимость: ${activeContent.style.display}`);
// 	}
// }
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

