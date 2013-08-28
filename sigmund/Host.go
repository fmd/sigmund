package sigmund

type Host struct {
	Redis   *RedisClient
	Name    string
	Conn    interface{}
	Data    interface{}
	Inputs  map[string]interface{}
	Outputs map[string]interface{}
}

func (h *Host) Save() (error, error) {
	var err error
	var removal_err error

	defer func() {
		if err != nil {
			removal_err = h.Remove()
		}
	}()

	if err = h.Redis.AddHost(h.Name); err != nil {
		return err, removal_err
	}

	if err = h.Redis.AddConn(h.Name, h.Conn); err != nil {
		return err, removal_err
	}

	if err = h.Redis.AddData(h.Name, h.Data); err != nil {
		return err, removal_err
	}

	if err = h.Redis.AddInputs(h.Name, h.Inputs); err != nil {
		return err, removal_err
	}

	if err = h.Redis.AddOutputs(h.Name, h.Outputs); err != nil {
		return err, removal_err
	}

	return err, removal_err
}

func (h *Host) Remove() error {
	var err error

	if err = h.Redis.RemoveOutputs(h.Name); err != nil {
		return err
	}

	if err = h.Redis.RemoveInputs(h.Name); err != nil {
		return err
	}

	if err = h.Redis.RemoveData(h.Name); err != nil {
		return err
	}

	if err = h.Redis.RemoveConn(h.Name); err != nil {
		return err
	}

	if err = h.Redis.RemoveHost(h.Name); err != nil {
		return err
	}

	return err
}
