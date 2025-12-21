#pragma once

#include <QDialog>
#include <QList>

#include "models.h"

class ApiClient;
class QTableWidget;
class QAction;

class LocationsDialog final : public QDialog {
    Q_OBJECT

public:
    explicit LocationsDialog(ApiClient* apiClient, QWidget* parent = nullptr);

private:
    void buildUi();
    void reload();
    qint64 selectedId() const;

    void addLocation();
    void editLocation();
    void deleteLocation();

    ApiClient* apiClient_;
    QList<LocationItem> locations_;

    QTableWidget* table_;
    QAction* addAction_;
    QAction* editAction_;
    QAction* deleteAction_;
    QAction* refreshAction_;
};
