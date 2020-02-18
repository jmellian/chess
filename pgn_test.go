package chess

import (
	"bufio"
	"github.com/notnil/chess"
	"os"
	"strings"
	"testing"
)

type pgnTest struct {
	PostPos *Position
	PGN     string
}

var (
	validPGNs = []pgnTest{
		{
			PostPos: unsafeFEN("4r3/6P1/2p2P1k/1p6/pP2p1R1/P1B5/2P2K2/3r4 b - - 0 45"),
			PGN: `[Event "?"]
            [Site "?"]
            [Date "1997.05.03"]
            [Round "1"]
            [White "Kasparov"]
            [Black "Deep-Blue"]
            [Result "1-0"]
            [WhiteElo "2795"]

            1. Nf3 d5 2. g3 Bg4 3. b3 Nd7 4. Bb2 e6 5. Bg2 Ngf6 6. O-O c6
            7. d3 Bd6 8. Nbd2 O-O 9. h3 Bh5 10. e3 h6 11. Qe1 Qa5 12. a3
            Bc7 13. Nh4 g5 14. Nhf3 e5 15. e4 Rfe8 16. Nh2 Qb6 17. Qc1 a5
            18. Re1 Bd6 19. Ndf1 dxe4 20. dxe4 Bc5 21. Ne3 Rad8 22. Nhf1 g4
            23. hxg4 Nxg4 24. f3 Nxe3 25. Nxe3 Be7 26. Kh1 Bg5 27. Re2 a4
            28. b4 f5 29. exf5 e4 30. f4 Bxe2 31. fxg5 Ne5 32. g6 Bf3 33. Bc3
            Qb5 34. Qf1 Qxf1+ 35. Rxf1 h5 36. Kg1 Kf8 37. Bh3 b5 38. Kf2 Kg7
            39. g4 Kh6 40. Rg1 hxg4 41. Bxg4 Bxg4 42. Nxg4+ Nxg4+ 43. Rxg4
            Rd5 44. f6 Rd1 45. g7 1-0`,
		},
		{
			PostPos: unsafeFEN("4r3/6P1/2p2P1k/1p6/pP2p1R1/P1B5/2P2K2/3r4 b - - 0 45"),
			PGN: `[Event "?"]
[Site "http://lichess.org/4HXJOtpN"]
[Date "1997.05.03"]
[White "Kasparov (2795)"]
[Black "Deep-Blue"]
[Result "1-0"]
[WhiteElo "?"]
[BlackElo "?"]
[PlyCount "89"]
[Variant "Standard"]
[TimeControl "-"]
[ECO "A07"]
[Opening "King's Indian Attack, General"]
[Termination "Normal"]
[Annotator "lichess.org"]

1. Nf3 d5 2. g3 { King's Indian Attack, General } Bg4 3. b3 Nd7 4. Bb2 e6 5. Bg2 Ngf6 6. O-O c6 7. d3 Bd6 8. Nbd2 O-O 9. h3 Bh5 10. e3 h6 11. Qe1 Qa5 12. a3 Bc7 13. Nh4 g5 14. Nhf3 e5 15. e4 Rfe8 16. Nh2 Qb6 17. Qc1 a5 18. Re1 Bd6 19. Ndf1 dxe4 20. dxe4 Bc5 21. Ne3 Rad8 22. Nhf1 g4 23. hxg4 Nxg4 24. f3 Nxe3 25. Nxe3 Be7 26. Kh1 Bg5 27. Re2 a4 28. b4 f5 29. exf5 e4 30. f4 Bxe2 31. fxg5 Ne5 32. g6 Bf3 33. Bc3 Qb5 34. Qf1 Qxf1+ 35. Rxf1 h5 36. Kg1 Kf8 37. Bh3 b5 38. Kf2 Kg7?! { (0.70 → 1.52) Inaccuracy. The best move was Rd6. } (38... Rd6 39. Re1 Re7 40. Rg1 Re8 41. g4 h4 42. g5 Kg8 43. Rf1 Kf8 44. Re1 Kg8 45. Rb1 Rdd8 46. Rf1 Kf8 47. Rg1 Kg8 48. Rb1) 39. g4 Kh6?! { (1.42 → 2.07) Inaccuracy. The best move was h4. } (39... h4 40. g5 Kf8 41. Bg2 Ng4+ 42. Nxg4 Bxg4 43. Ke3 h3 44. Rf4 hxg2 45. Rxg4 Rd1 46. Rxg2 Rf1 47. Rf2 Rg1 48. Bf6 Rh1 49. Kd4) 40. Rg1 hxg4 41. Bxg4 Bxg4 42. Nxg4+ Nxg4+ 43. Rxg4 Rd5 44. f6 Rd1?? { (1.60 → 8.36) Blunder. The best move was Rf5+. } (44... Rf5+ 45. Ke2 Rg8 46. g7 e3 47. Bd4 Kh7 48. Kxe3 Re8+ 49. Re4 Rxe4+ 50. Kxe4 Rf1 51. Ke5 Kg8 52. Ke6 Rf4 53. c3 Rf1 54. Ke7) 45. g7 { Black resigns } 1-0`,
		},
	}
)

func TestValidPGNs(t *testing.T) {
	for _, test := range validPGNs {
		game, err := decodePGN(test.PGN)
		if err != nil {
			t.Fatalf("recieved unexpected pgn error %s", err.Error())
		}
		if game.Position().String() != test.PostPos.String() {
			t.Fatalf("expected board to be \n%s\nFEN:%s\n but got \n%s\n\nFEN:%s\n",
				test.PostPos.board.Draw(), test.PostPos.String(),
				game.Position().board.Draw(), game.Position().String())
		}
	}
}

func TestLimitedGamesFromPGNReader(t *testing.T) {
	inputPath := "./assets/twic0920.pgn"
	in, err := os.Open(inputPath)
	if err != nil {
		t.Fatalf("unexpected pgn read error %s", err.Error())
	}
	defer in.Close()

	readNrGames := 1000
	processed_games := 0
	tel := 0
	br := bufio.NewReader(in)
	for br != nil {
		_, br, processed_games, err = LimitedGamesFromPGNReader(br, readNrGames)
		if err != nil {
			t.Fatalf("unexpected pgn read processing games error %s", err.Error())
		}
		tel = tel + processed_games
	}
	if tel != 2245 {
		t.Fatalf("2245 games should have been read from ./assets/twic0920.pgn. %d games read", processed_games)
	}
}

func TestGamesFromPGN(t *testing.T) {
	inputPath := "./assets/twic0920.pgn"
	in, err := os.Open(inputPath)
	if err != nil {
		t.Fatalf("unexpected pgn read error %s", err.Error())
	}
	defer in.Close()

	_, err = chess.GamesFromPGN(in)
	if err != nil {
		t.Fatalf("Error reading ./assets/twic0920.pgn")
	}
}

func BenchmarkPGN(b *testing.B) {
	pgn := `[Event "?"]
	[Site "?"]
	[Date "1997.05.03"]
	[Round "1"]
	[White "Kasparov"]
	[Black "Deep-Blue"]
	[Result "1-0"]
	[WhiteElo "2795"]

	1. Nf3 d5 2. g3 Bg4 3. b3 Nd7 4. Bb2 e6 5. Bg2 Ngf6 6. O-O c6
	7. d3 Bd6 8. Nbd2 O-O 9. h3 Bh5 10. e3 h6 11. Qe1 Qa5 12. a3
	Bc7 13. Nh4 g5 14. Nhf3 e5 15. e4 Rfe8 16. Nh2 Qb6 17. Qc1 a5
	18. Re1 Bd6 19. Ndf1 dxe4 20. dxe4 Bc5 21. Ne3 Rad8 22. Nhf1 g4
	23. hxg4 Nxg4 24. f3 Nxe3 25. Nxe3 Be7 26. Kh1 Bg5 27. Re2 a4
	28. b4 f5 29. exf5 e4 30. f4 Bxe2 31. fxg5 Ne5 32. g6 Bf3 33. Bc3
	Qb5 34. Qf1 Qxf1+ 35. Rxf1 h5 36. Kg1 Kf8 37. Bh3 b5 38. Kf2 Kg7
	39. g4 Kh6 40. Rg1 hxg4 41. Bxg4 Bxg4 42. Nxg4+ Nxg4+ 43. Rxg4
	Rd5 44. f6 Rd1 45. g7 1-0`

	for n := 0; n < b.N; n++ {
		opt, _ := PGN(strings.NewReader(pgn))
		NewGame(opt)
	}
}
