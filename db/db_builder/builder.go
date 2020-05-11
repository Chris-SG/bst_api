package db_builder

import (
	"github.com/chris-sg/bst_api/models/api_models"
	"github.com/chris-sg/bst_api/models/bst_models"
	"github.com/chris-sg/bst_api/models/ddr_models"
	"github.com/chris-sg/bst_api/models/drs_models"
	"github.com/chris-sg/bst_api/models/user_models"
	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
)

type DbMigrator interface {
	Create()
	CreateTables()
	CreateConstraints()

	createUserTables()
	createApiTables()
	createBstTables()
	createDdrTables()
	createDrsTables()
	createDdrConstraints()
	createDrsConstraints()
}

func CreateDbMigratorPostgres(db *gorm.DB) DbMigratorPostgres {
	return DbMigratorPostgres{db}
}

type DbMigratorPostgres struct {
	db *gorm.DB
}

func (migrator DbMigratorPostgres) Create() {
	glog.Infoln("creating db")
	migrator.CreateTables()
	migrator.CreateConstraints()
}

func (migrator DbMigratorPostgres) CreateTables() {
	glog.Infoln("creating db tables")
	migrator.createUserTables()
	migrator.createApiTables()
	migrator.createBstTables()
	migrator.createDdrTables()
	migrator.createDrsTables()
}

func (migrator DbMigratorPostgres) CreateConstraints() {
	glog.Infoln("creating db constraints")
	migrator.createDdrConstraints()
	migrator.createDrsConstraints()
}

func (migrator DbMigratorPostgres) createUserTables() {
	errs := migrator.db.AutoMigrate(&user_models.User{}).GetErrors()
	if errs != nil && len(errs) > 0 {
		glog.Warningln("automigration for user table user_models.User contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}
}

func (migrator DbMigratorPostgres) createApiTables() {
	errs := migrator.db.AutoMigrate(&api_models.AutomaticJob{}).GetErrors()

	if errs != nil && len(errs) > 0 {
		glog.Warningln("automigration for api table api_models.AutomaticJob contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}
}

func (migrator DbMigratorPostgres) createBstTables() {
	errs := migrator.db.AutoMigrate(&bst_models.BstProfile{}).GetErrors()

	if errs != nil && len(errs) > 0 {
		glog.Warningln("automigration for bst table bst_models.BstProfile contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}
}

func (migrator DbMigratorPostgres) createDdrTables() {
	errs := migrator.db.AutoMigrate(&ddr_models.Song{}, &ddr_models.SongDifficulty{},
						  &ddr_models.PlayerDetails{}, &ddr_models.Playcount{},
						  &ddr_models.Score{}, ddr_models.SongStatistics{},
						  &ddr_models.WorkoutData{}).
			  GetErrors()
	if errs != nil && len(errs) > 0 {
		glog.Warningln("automigration for ddr tables contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}
}

func (migrator DbMigratorPostgres) createDdrConstraints() {
	errs := migrator.db.Model(&ddr_models.SongDifficulty{}).
		AddForeignKey("song_id", "public.\"ddrSongs\"(id)", "CASCADE", "CASCADE").
		GetErrors()
	if errs != nil && len(errs) > 0 {
		glog.Warningln("fk creation for ddr_models.SongDifficulty contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}

	errs = migrator.db.Model(&ddr_models.PlayerDetails{}).
		      AddForeignKey("eagate_user", "public.\"eaGateUser\"(account_name)", "RESTRICT", "RESTRICT").
			  GetErrors()
	if errs != nil && len(errs) > 0 {
		glog.Warningln("fk creation for ddr_models.PlayerDetails contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}

	errs = migrator.db.Model(&ddr_models.Playcount{}).
		AddForeignKey("player_code", "public.\"ddrPlayerDetails\"(code)", "RESTRICT", "RESTRICT").
		GetErrors()
	if errs != nil && len(errs) > 0 {
		glog.Warningln("fk creation for ddr_models.Playcount contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}

	errs = migrator.db.Model(&ddr_models.WorkoutData{}).
		AddForeignKey("player_code", "public.\"ddrPlayerDetails\"(code)", "RESTRICT", "RESTRICT").
		GetErrors()
	if errs != nil && len(errs) > 0 {
		glog.Warningln("fk creation for ddr_models.WorkoutData contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}

	errs = migrator.db.Model(&ddr_models.SongStatistics{}).
		AddForeignKey("song_id,mode,difficulty", "public.\"ddrSongDifficulties\"(song_id,mode,difficulty)", "RESTRICT", "RESTRICT").
		GetErrors()
	if errs != nil && len(errs) > 0 {
		glog.Warningln("fk creation for ddr_models.SongStatistics contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}

	errs = migrator.db.Model(&ddr_models.SongStatistics{}).
		AddForeignKey("player_code", "public.\"ddrPlayerDetails\"(code)", "RESTRICT", "RESTRICT").
		GetErrors()
	if errs != nil && len(errs) > 0 {
		glog.Warningln("fk creation for ddr_models.SongStatistics contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}

	errs = migrator.db.Model(&ddr_models.Score{}).
		AddForeignKey("song_id,mode,difficulty", "public.\"ddrSongDifficulties\"(song_id,mode,difficulty)", "RESTRICT", "RESTRICT").
		GetErrors()
	if errs != nil && len(errs) > 0 {
		glog.Warningln("fk creation for ddr_models.Score contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}

	errs = migrator.db.Model(&ddr_models.Score{}).
		AddForeignKey("player_code", "public.\"ddrPlayerDetails\"(code)", "RESTRICT", "RESTRICT").
		GetErrors()
	if errs != nil && len(errs) > 0 {
		glog.Warningln("fk creation for ddr_models.Score contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}
}

func (migrator DbMigratorPostgres) createDrsTables() {
	errs := migrator.db.AutoMigrate(&drs_models.PlayerDetails{}, &drs_models.Difficulty{},
		&drs_models.DancerInfo{}, &drs_models.PlayerScore{},
		&drs_models.PlayerSongStats{}, &drs_models.Song{},
		&drs_models.PlayerProfileSnapshot{}, &drs_models.PlayHist{},
		&drs_models.MusicData{}).
		GetErrors()
	if errs != nil && len(errs) > 0 {
		glog.Warningln("automigration for drs tables contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}
}

func (migrator DbMigratorPostgres) createDrsConstraints() {
	errs := migrator.db.Model(&drs_models.PlayerDetails{}).
		AddForeignKey("eagate_user", "public.\"eaGateUser\"(account_name)", "RESTRICT", "RESTRICT").
		GetErrors()
	if errs != nil && len(errs) > 0 {
		glog.Warningln("fk creation for drs_models.PlayerDetails contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}

	errs = migrator.db.Model(&drs_models.PlayerProfileSnapshot{}).
		AddForeignKey("player_code", "public.\"drsPlayerDetails\"(code)", "RESTRICT", "RESTRICT").
		GetErrors()
	if errs != nil && len(errs) > 0 {
		glog.Warningln("fk creation for drs_models.PlayerProfileSnapshot contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}

	errs = migrator.db.Model(&drs_models.Difficulty{}).
		AddForeignKey("song_id", "public.\"drsSongs\"(song_id)", "CASCADE", "CASCADE").
		GetErrors()
	if errs != nil && len(errs) > 0 {
		glog.Warningln("fk creation for drs_models.Difficulty contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}

	errs = migrator.db.Model(&drs_models.PlayerSongStats{}).
		AddForeignKey("player_code", "public.\"drsPlayerDetails\"(code)", "RESTRICT", "RESTRICT").
		AddForeignKey("song_id", "public.\"drsSongs\"(song_id)", "CASCADE", "CASCADE").
		AddForeignKey("mode", "public.\"drsDifficulties\"(mode)", "RESTRICT", "RESTRICT").
		AddForeignKey("difficulty", "public.\"drsDifficulties\"(difficulty)", "RESTRICT", "RESTRICT").
		GetErrors()
	if errs != nil && len(errs) > 0 {
		glog.Warningln("fk creation for drs_models.PlayerSongStats contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}

	errs = migrator.db.Model(&drs_models.PlayerScore{}).
		AddForeignKey("player_code", "public.\"drsPlayerDetails\"(code)", "RESTRICT", "RESTRICT").
		AddForeignKey("song_id", "public.\"drsSongs\"(song_id)", "CASCADE", "CASCADE").
		AddForeignKey("mode", "public.\"drsDifficulties\"(mode)", "RESTRICT", "RESTRICT").
		AddForeignKey("difficulty", "public.\"drsDifficulties\"(difficulty)", "RESTRICT", "RESTRICT").
		GetErrors()
	if errs != nil && len(errs) > 0 {
		glog.Warningln("fk creation for drs_models.PlayerScore contained errors:")
		for _, err := range errs {
			glog.Warningf("\t%s\n", err.Error())
		}
	}
}