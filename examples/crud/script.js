(function() {
    window.crudCreateModal = {
        open: function() {
            const modal = document.getElementById('crud-create-modal');
            if (modal) modal.style.display = 'flex';
        },
        close: function() {
            const modal = document.getElementById('crud-create-modal');
            if (modal) modal.style.display = 'none';
        }
    };

    window.crudEditModal = {
        open: function(id, name, email, role) {
            const modal = document.getElementById('crud-edit-modal');
            const idEl = document.getElementById('crud-edit-id');
            const nameEl = document.getElementById('crud-edit-name');
            const emailEl = document.getElementById('crud-edit-email');
            const roleEl = document.getElementById('crud-edit-role');
            if (modal && idEl && nameEl && emailEl && roleEl) {
                idEl.value = id;
                nameEl.value = name;
                emailEl.value = email;
                roleEl.value = role;
                modal.style.display = 'flex';
            }
        },
        close: function() {
            const modal = document.getElementById('crud-edit-modal');
            if (modal) modal.style.display = 'none';
        }
    };

    window.crudDeleteModal = {
        open: function(id, name) {
            const modal = document.getElementById('crud-delete-modal');
            const idEl = document.getElementById('crud-delete-id');
            const nameEl = document.getElementById('crud-delete-name');
            if (modal && idEl && nameEl) {
                idEl.value = id;
                nameEl.textContent = name;
                modal.style.display = 'flex';
            }
        },
        close: function() {
            const modal = document.getElementById('crud-delete-modal');
            if (modal) modal.style.display = 'none';
        }
    };
})();
