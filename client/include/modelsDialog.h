#pragma once

#include <QDialog>
#include <QList>

#include "models.h"

class ApiClient;
class QTableWidget;
class QAction;

class ModelsDialog final : public QDialog {
    Q_OBJECT

public:
    explicit ModelsDialog(ApiClient* apiClient, QWidget* parent = nullptr);

private:
    void buildUi();
    void reload();
    qint64 selectedId() const;

    void addModel();
    void editModel();
    void deleteModel();

    ApiClient* apiClient_;
    QList<ModelItem> models_;

    QTableWidget* table_;
    QAction* addAction_;
    QAction* editAction_;
    QAction* deleteAction_;
    QAction* refreshAction_;
};
