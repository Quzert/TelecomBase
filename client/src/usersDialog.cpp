#include "usersDialog.h"

#include "apiClient.h"

#include <QAction>
#include <QDialogButtonBox>
#include <QMessageBox>
#include <QPushButton>
#include <QTableWidget>
#include <QToolBar>
#include <QVBoxLayout>
#include <QHeaderView>
#include <QStyle>

#include <QIcon>

#include "messageBoxUtils.h"

UsersDialog::UsersDialog(ApiClient* apiClient, QWidget* parent)
    : QDialog(parent)
    , apiClient_(apiClient)
    , users_()
    , table_(nullptr)
    , approveAction_(nullptr)
    , disableAction_(nullptr)
    , deleteAction_(nullptr)
    , refreshAction_(nullptr) {
    setWindowTitle("Пользователи");
    buildUi();
    reload();
}

void UsersDialog::buildUi() {
    QVBoxLayout* root = new QVBoxLayout(this);
    root->setContentsMargins(16, 16, 16, 16);
    root->setSpacing(12);

    QToolBar* toolbar = new QToolBar(this);
    toolbar->setMovable(false);
    toolbar->setToolButtonStyle(Qt::ToolButtonIconOnly);
    approveAction_ = toolbar->addAction("Выдать доступ");
    disableAction_ = toolbar->addAction("Забрать доступ");
    deleteAction_ = toolbar->addAction("Удалить аккаунт");
    toolbar->addSeparator();
    refreshAction_ = toolbar->addAction("Обновить");
    root->addWidget(toolbar);

    // Используем стандартные иконки Qt (без своих ресурсов).
    if (approveAction_) {
        approveAction_->setIcon(style()->standardIcon(QStyle::SP_DialogApplyButton));
        approveAction_->setToolTip("Подтвердить пользователя");
        approveAction_->setText(QString());
    }
    if (disableAction_) {
        disableAction_->setIcon(style()->standardIcon(QStyle::SP_DialogCancelButton));
        disableAction_->setToolTip("Отклонить/забрать доступ");
        disableAction_->setText(QString());
    }

    table_ = new QTableWidget(0, 5, this);
    table_->setHorizontalHeaderLabels({"ID", "Логин", "Роль", "Доступ", "Создан"});
    table_->setSelectionBehavior(QAbstractItemView::SelectRows);
    table_->setSelectionMode(QAbstractItemView::SingleSelection);
    table_->setEditTriggers(QAbstractItemView::NoEditTriggers);
    table_->setColumnHidden(0, true);
    table_->setAlternatingRowColors(true);
    table_->setShowGrid(false);
    table_->verticalHeader()->setVisible(false);
    table_->horizontalHeader()->setStretchLastSection(true);
    root->addWidget(table_);

    QDialogButtonBox* closeBox = new QDialogButtonBox(QDialogButtonBox::Close, this);
    if (auto* btn = closeBox->button(QDialogButtonBox::Close)) {
        btn->setIcon(QIcon());
    }
    connect(closeBox, &QDialogButtonBox::rejected, this, &QDialog::reject);
    connect(closeBox, &QDialogButtonBox::accepted, this, &QDialog::accept);
    root->addWidget(closeBox);

    connect(refreshAction_, &QAction::triggered, this, &UsersDialog::reload);
    connect(approveAction_, &QAction::triggered, this, &UsersDialog::approveUser);
    connect(disableAction_, &QAction::triggered, this, &UsersDialog::disableUser);
    connect(deleteAction_, &QAction::triggered, this, &UsersDialog::deleteUser);

    resize(780, 440);
}

void UsersDialog::reload() {
    if (!apiClient_) {
        return;
    }

    QString err;
    QList<UserItem> list;
    if (!apiClient_->listUsers(list, err)) {
        UiUtils::warning(this, "Ошибка", err.isEmpty() ? "Не удалось загрузить пользователей" : err);
        return;
    }

    users_ = list;
    table_->setRowCount(0);
    table_->setRowCount(users_.size());

    for (int i = 0; i < users_.size(); ++i) {
        const auto& u = users_[i];
        table_->setItem(i, 0, new QTableWidgetItem(QString::number(u.id)));
        table_->setItem(i, 1, new QTableWidgetItem(u.username));
        table_->setItem(i, 2, new QTableWidgetItem(u.role));
        table_->setItem(i, 3, new QTableWidgetItem(u.approved ? "Да" : "Нет"));
        table_->setItem(i, 4, new QTableWidgetItem(u.createdAt));
    }
}

qint64 UsersDialog::selectedId() const {
    if (!table_) {
        return 0;
    }
    const auto ranges = table_->selectedRanges();
    if (ranges.isEmpty()) {
        return 0;
    }
    const int row = ranges.first().topRow();
    const QTableWidgetItem* item = table_->item(row, 0);
    if (!item) {
        return 0;
    }
    return item->text().toLongLong();
}

void UsersDialog::approveUser() {
    if (!apiClient_) {
        return;
    }

    const qint64 id = selectedId();
    if (id <= 0) {
        UiUtils::information(this, "Доступ", "Выберите пользователя");
        return;
    }

    QString err;
    if (!apiClient_->setUserApproved(id, true, err)) {
        UiUtils::warning(this, "Не удалось выдать доступ", err.isEmpty() ? "Ошибка" : err);
        return;
    }

    reload();
}

void UsersDialog::disableUser() {
    if (!apiClient_) {
        return;
    }

    const qint64 id = selectedId();
    if (id <= 0) {
        UiUtils::information(this, "Доступ", "Выберите пользователя");
        return;
    }

    const auto answer = UiUtils::question(this, "Доступ", "Забрать доступ у выбранного пользователя?");
    if (answer != QMessageBox::Yes) {
        return;
    }

    QString err;
    if (!apiClient_->setUserApproved(id, false, err)) {
        UiUtils::warning(this, "Не удалось забрать доступ", err.isEmpty() ? "Ошибка" : err);
        return;
    }

    reload();
}

void UsersDialog::deleteUser() {
    if (!apiClient_) {
        return;
    }

    const qint64 id = selectedId();
    if (id <= 0) {
        UiUtils::information(this, "Удаление", "Выберите пользователя");
        return;
    }

    const auto answer = UiUtils::question(this, "Удаление", "Удалить выбранный аккаунт?");
    if (answer != QMessageBox::Yes) {
        return;
    }

    QString err;
    if (!apiClient_->deleteUser(id, err)) {
        UiUtils::warning(this, "Не удалось удалить", err.isEmpty() ? "Ошибка" : err);
        return;
    }

    reload();
}
