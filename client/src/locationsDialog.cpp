#include "locationsDialog.h"

#include "apiClient.h"

#include <QAction>
#include <QDialogButtonBox>
#include <QFormLayout>
#include <QLineEdit>
#include <QMessageBox>
#include <QPushButton>
#include <QTableWidget>
#include <QToolBar>
#include <QVBoxLayout>
#include <QHeaderView>

#include "messageBoxUtils.h"

#include <QIcon>

static bool editLocationDialog(QWidget* parent, const QString& title, QString& ioName, QString& ioNote) {
    QDialog dlg(parent);
    dlg.setWindowTitle(title);

    QVBoxLayout* root = new QVBoxLayout(&dlg);
    root->setContentsMargins(16, 16, 16, 16);
    root->setSpacing(12);
    QFormLayout* form = new QFormLayout();

    QLineEdit* nameEdit = new QLineEdit(&dlg);
    nameEdit->setText(ioName);
    form->addRow("Название:", nameEdit);

    QLineEdit* noteEdit = new QLineEdit(&dlg);
    noteEdit->setText(ioNote);
    form->addRow("Примечание:", noteEdit);

    root->addLayout(form);

    QDialogButtonBox* buttons = new QDialogButtonBox(QDialogButtonBox::Ok | QDialogButtonBox::Cancel, &dlg);
    // Принудительно убираем иконки (некоторые темы добавляют их автоматически)
    if (auto* ok = buttons->button(QDialogButtonBox::Ok)) {
        ok->setIcon(QIcon());
    }
    if (auto* cancel = buttons->button(QDialogButtonBox::Cancel)) {
        cancel->setIcon(QIcon());
    }
    root->addWidget(buttons);
    QObject::connect(buttons, &QDialogButtonBox::accepted, &dlg, &QDialog::accept);
    QObject::connect(buttons, &QDialogButtonBox::rejected, &dlg, &QDialog::reject);

    if (dlg.exec() != QDialog::Accepted) {
        return false;
    }

    ioName = nameEdit->text().trimmed();
    ioNote = noteEdit->text().trimmed();
    if (ioName.isEmpty()) {
        UiUtils::information(parent, title, "Название обязательно");
        return false;
    }

    return true;
}

LocationsDialog::LocationsDialog(ApiClient* apiClient, QWidget* parent)
    : QDialog(parent)
    , apiClient_(apiClient)
    , locations_()
    , table_(nullptr)
    , addAction_(nullptr)
    , editAction_(nullptr)
    , deleteAction_(nullptr)
    , refreshAction_(nullptr) {
    setWindowTitle("Справочник: локации");
    buildUi();
    reload();
}

void LocationsDialog::buildUi() {
    QVBoxLayout* root = new QVBoxLayout(this);
    root->setContentsMargins(16, 16, 16, 16);
    root->setSpacing(12);

    QToolBar* toolbar = new QToolBar(this);
    toolbar->setMovable(false);
    addAction_ = toolbar->addAction("Добавить");
    editAction_ = toolbar->addAction("Редактировать");
    deleteAction_ = toolbar->addAction("Удалить");
    toolbar->addSeparator();
    refreshAction_ = toolbar->addAction("Обновить");
    root->addWidget(toolbar);

    table_ = new QTableWidget(0, 3, this);
    table_->setHorizontalHeaderLabels({"ID", "Название", "Примечание"});
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

    connect(refreshAction_, &QAction::triggered, this, &LocationsDialog::reload);
    connect(addAction_, &QAction::triggered, this, &LocationsDialog::addLocation);
    connect(editAction_, &QAction::triggered, this, &LocationsDialog::editLocation);
    connect(deleteAction_, &QAction::triggered, this, &LocationsDialog::deleteLocation);

    resize(820, 420);
}

void LocationsDialog::reload() {
    if (!apiClient_) {
        return;
    }

    QString err;
    QList<LocationItem> list;
    if (!apiClient_->listLocations(list, err)) {
        UiUtils::warning(this, "Ошибка", err.isEmpty() ? "Не удалось загрузить локации" : err);
        return;
    }

    locations_ = list;
    table_->setRowCount(0);
    table_->setRowCount(locations_.size());

    for (int i = 0; i < locations_.size(); ++i) {
        const auto& l = locations_[i];
        table_->setItem(i, 0, new QTableWidgetItem(QString::number(l.id)));
        table_->setItem(i, 1, new QTableWidgetItem(l.name));
        table_->setItem(i, 2, new QTableWidgetItem(l.note));
    }
}

qint64 LocationsDialog::selectedId() const {
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

void LocationsDialog::addLocation() {
    if (!apiClient_) {
        return;
    }

    QString name;
    QString note;
    if (!editLocationDialog(this, "Добавить локацию", name, note)) {
        return;
    }

    QString err;
    if (!apiClient_->createLocation(name, note, err)) {
        UiUtils::warning(this, "Не удалось добавить", err.isEmpty() ? "Ошибка" : err);
        return;
    }

    reload();
}

void LocationsDialog::editLocation() {
    if (!apiClient_) {
        return;
    }

    const qint64 id = selectedId();
    if (id <= 0) {
        UiUtils::information(this, "Редактирование", "Выберите запись");
        return;
    }

    QString name;
    QString note;
    for (const auto& l : locations_) {
        if (l.id == id) {
            name = l.name;
            note = l.note;
            break;
        }
    }

    if (!editLocationDialog(this, "Редактировать локацию", name, note)) {
        return;
    }

    QString err;
    if (!apiClient_->updateLocation(id, name, note, err)) {
        UiUtils::warning(this, "Не удалось сохранить", err.isEmpty() ? "Ошибка" : err);
        return;
    }

    reload();
}

void LocationsDialog::deleteLocation() {
    if (!apiClient_) {
        return;
    }

    const qint64 id = selectedId();
    if (id <= 0) {
        UiUtils::information(this, "Удаление", "Выберите запись");
        return;
    }

    const auto answer = UiUtils::question(this, "Удаление", "Удалить выбранную локацию?");
    if (answer != QMessageBox::Yes) {
        return;
    }

    QString err;
    if (!apiClient_->deleteLocation(id, err)) {
        UiUtils::warning(this, "Не удалось удалить", err.isEmpty() ? "Ошибка" : err);
        return;
    }

    reload();
}
