package gradingservice

import (
	"context"
	"fmt"
	"log"
	Pb "rubr/proto/grade"
	"strconv"
	"strings"
)

func (s *Server) GetCriteriaMarks(ctx context.Context, req *Pb.GetCriteriaMarksRequest) (*Pb.GetCriteriaMarksResponse, error) {
	log.Printf("Получен запрос GetCriteriaMarks для work_id: %d", req.WorkId)

	// Проверка входных данных
	if req.WorkId <= 0 {
		log.Printf("Неверный work_id: %d", req.WorkId)
		return &Pb.GetCriteriaMarksResponse{Error: "work_id должен быть положительным"}, nil
	}

	// SQL-запрос для получения оценок
	query := `
        SELECT criteria_id, mark, COALESCE(comment, '')
        FROM student_criteria_marks
        WHERE student_work_id = $1`
	rows, err := s.Db.QueryContext(ctx, query, req.WorkId)
	if err != nil {
		log.Printf("Ошибка выполнения запроса для work_id %d: %v", req.WorkId, err)
		return &Pb.GetCriteriaMarksResponse{Error: fmt.Sprintf("ошибка базы данных: %v", err)}, nil
	}
	defer rows.Close()

	// Сбор результатов
	var marks []*Pb.CriterionMark
	for rows.Next() {
		var criterionID int64
		var mark float32
		var comment string
		if err := rows.Scan(&criterionID, &mark, &comment); err != nil {
			log.Printf("Ошибка сканирования строки для work_id %d: %v", req.WorkId, err)
			return &Pb.GetCriteriaMarksResponse{Error: fmt.Sprintf("ошибка обработки данных: %v", err)}, nil
		}
		marks = append(marks, &Pb.CriterionMark{
			CriterionId: int32(criterionID),
			Mark:        mark,
			Comment:     comment,
		})
	}

	// Проверка ошибок после итерации
	if err := rows.Err(); err != nil {
		log.Printf("Ошибка итерации строк для work_id %d: %v", req.WorkId, err)
		return &Pb.GetCriteriaMarksResponse{Error: fmt.Sprintf("ошибка обработки данных: %v", err)}, nil
	}

	log.Printf("Найдено %d оценок для work_id %d", len(marks), req.WorkId)
	return &Pb.GetCriteriaMarksResponse{
		Marks: marks,
	}, nil
}

func (s *Server) ListSubjects(ctx context.Context, req *Pb.ListSubjectsRequest) (*Pb.ListSubjectsResponse, error) {
	query := `SELECT name, grades, average FROM student_subjects WHERE student_id = $1`
	rows, err := s.Db.QueryContext(ctx, query, req.StudentId)
	if err != nil {
		return &Pb.ListSubjectsResponse{Error: err.Error()}, nil
	}
	defer rows.Close()

	var subjects []*Pb.Subject
	for rows.Next() {
		var subject Pb.Subject
		var gradesStr string
		if err := rows.Scan(&subject.Name, &gradesStr, &subject.Average); err != nil {
			return &Pb.ListSubjectsResponse{Error: err.Error()}, nil
		}
		// Парсинг grades (предполагается, что grades хранится как строка, например, "4.0,3.5,4.5")
		for _, g := range strings.Split(gradesStr, ",") {
			grade, _ := strconv.ParseFloat(g, 32)
			subject.Grades = append(subject.Grades, float32(grade))
		}
		subjects = append(subjects, &subject)
	}
	if err := rows.Err(); err != nil {
		return &Pb.ListSubjectsResponse{Error: err.Error()}, nil
	}
	return &Pb.ListSubjectsResponse{Subjects: subjects}, nil
}
